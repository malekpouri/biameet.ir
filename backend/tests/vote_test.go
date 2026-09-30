package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"sync"
	"testing"

	"biameet.ir/models"
)

func vote(name, password, token string, slots ...string) models.VoteRequest {
	req := models.VoteRequest{VoterName: name, Password: password, Token: token, Votes: []models.VoteItem{}}
	for _, s := range slots {
		req.Votes = append(req.Votes, models.VoteItem{TimeslotID: s})
	}
	return req
}

func TestVoteWithoutPasswordCannotBeEditedByAnyone(t *testing.T) {
	e := newEnv(t)
	id, slots := createFixed(t, e)
	path := "/api/v1/sessions/" + id + "/vote"

	r := e.do(t, "POST", path, vote("Ali", "", "", slots[0]))
	expectStatus(t, r, 200)
	if r.JSON(t)["token"] != nil {
		t.Fatal("no token should be issued without a password")
	}

	expectError(t, e.do(t, "POST", path, vote("Ali", "", "", slots[1])), 409, "name_taken_no_password")
	expectError(t, e.do(t, "POST", path, vote("Ali", "guess", "", slots[1])), 409, "name_taken_no_password")
}

func TestVotePasswordAndTokenFlow(t *testing.T) {
	e := newEnv(t)
	id, slots := createFixed(t, e)
	path := "/api/v1/sessions/" + id + "/vote"

	r := e.do(t, "POST", path, vote("Ali", "secret", "", slots[0]))
	expectStatus(t, r, 200)
	aliToken, _ := r.JSON(t)["token"].(string)
	if aliToken == "" {
		t.Fatal("expected a token")
	}
	expectStatus(t, e.do(t, "POST", path, vote("Sara", "other", "", slots[0])), 200)

	expectError(t, e.do(t, "POST", path, vote("Ali", "", "", slots[1])), 401, "password_required")
	expectError(t, e.do(t, "POST", path, vote("Ali", "wrong", "", slots[1])), 401, "invalid_password")
	expectError(t, e.do(t, "POST", path, vote("Ali", "", "forged", slots[1])), 401, "password_required")

	// The token stands in for the password…
	expectStatus(t, e.do(t, "POST", path, vote("Ali", "", aliToken, slots[1])), 200)
	// …but only for the participant it was issued to.
	expectError(t, e.do(t, "POST", path, vote("Sara", "", aliToken, slots[1])), 401, "password_required")

	s := getSession(t, e, id)
	for _, ts := range s.Timeslots {
		for _, v := range ts.Votes {
			if v.VoterName == "Ali" && ts.ID != slots[1] {
				t.Fatal("Ali's old vote should have been replaced")
			}
		}
	}

	// An empty list withdraws all votes.
	expectStatus(t, e.do(t, "POST", path, vote("Ali", "secret", "")), 200)
	for _, ts := range getSession(t, e, id).Timeslots {
		for _, v := range ts.Votes {
			if v.VoterName == "Ali" {
				t.Fatal("Ali's votes should be withdrawn")
			}
		}
	}
}

func TestVoteRejectsForeignTimeslot(t *testing.T) {
	e := newEnv(t)
	id, _ := createFixed(t, e)
	_, otherSlots := createFixed(t, e)
	expectError(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/vote", vote("Ali", "", "", otherSlots[0])), 400, "invalid_input")
	// Nothing was written: the name is still free.
	expectStatus(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/vote", vote("Ali", "", "")), 200)
}

func TestVoteOnUnknownSession(t *testing.T) {
	e := newEnv(t)
	expectError(t, e.do(t, "POST", "/api/v1/sessions/zzzzz/vote", vote("Ali", "", "")), 404, "session_not_found")
}

func TestVoteNameIsTrimmedAndRequired(t *testing.T) {
	e := newEnv(t)
	id, slots := createFixed(t, e)
	path := "/api/v1/sessions/" + id + "/vote"
	expectError(t, e.do(t, "POST", path, vote("   ", "", "", slots[0])), 400, "invalid_input")
	expectStatus(t, e.do(t, "POST", path, vote(" Ali ", "", "", slots[0])), 200)
	expectError(t, e.do(t, "POST", path, vote("Ali", "", "", slots[0])), 409, "name_taken_no_password")
}

// With WAL, busy_timeout and immediate transactions, simultaneous voters
// queue for the write lock instead of failing with SQLITE_BUSY.
func TestConcurrentVotes(t *testing.T) {
	e := newEnv(t)
	id, slots := createFixed(t, e)

	const n = 30
	errs := make(chan string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body, _ := json.Marshal(vote(fmt.Sprintf("voter-%d", i), "", "", slots...))
			req := httptest.NewRequest("POST", "/api/v1/sessions/"+id+"/vote", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := e.app.Test(req, -1)
			if err != nil {
				errs <- err.Error()
				return
			}
			if resp.StatusCode != 200 {
				b, _ := io.ReadAll(resp.Body)
				errs <- fmt.Sprintf("status %d: %s", resp.StatusCode, b)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if got := len(getSession(t, e, id).Timeslots[0].Votes); got != n {
		t.Fatalf("got %d votes, want %d", got, n)
	}
}
