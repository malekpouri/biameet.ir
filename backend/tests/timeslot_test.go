package tests

import (
	"testing"

	"biameet.ir/models"
)

func slot(start, end, by, password, token string) models.TimeslotRequest {
	return models.TimeslotRequest{StartUTC: start, EndUTC: end, CreatedBy: by, Password: password, Token: token}
}

const (
	s10 = "2030-01-01T10:00:00.000Z"
	s11 = "2030-01-01T11:00:00.000Z"
	s12 = "2030-01-01T12:00:00.000Z"
)

func TestAddTimeslotToUnknownSession(t *testing.T) {
	e := newEnv(t)
	expectError(t, e.do(t, "POST", "/api/v1/sessions/zzzzz/timeslots", slot(s10, s11, "Ali", "", "")), 404, "session_not_found")
}

func TestAddTimeslotToFixedSessionIsRejected(t *testing.T) {
	e := newEnv(t)
	id, _ := createFixed(t, e)
	expectError(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/timeslots", slot(s10, s11, "Ali", "", "")), 400, "fixed_session")
}

func TestAddTimeslotRangeAndDuplicates(t *testing.T) {
	e := newEnv(t)
	id := createDynamic(t, e)
	path := "/api/v1/sessions/" + id + "/timeslots"

	expectError(t, e.do(t, "POST", path, slot("2030-02-01T10:00:00Z", "2030-02-01T11:00:00Z", "Ali", "pw", "")), 400, "out_of_range")
	expectError(t, e.do(t, "POST", path, slot(s11, s10, "Ali", "pw", "")), 400, "invalid_input")

	r := e.do(t, "POST", path, slot(s10, s11, "Ali", "pw", ""))
	expectStatus(t, r, 201)
	if r.JSON(t)["token"] == "" {
		t.Fatal("expected a token for a password-protected proposer")
	}
	// Same slot written with a different ISO shape is still a duplicate.
	expectError(t, e.do(t, "POST", path, slot("2030-01-01T10:00:00Z", "2030-01-01T11:00:00Z", "Ali", "pw", "")), 409, "duplicate_timeslot")

	s := getSession(t, e, id)
	if len(s.Timeslots) != 1 || len(s.Timeslots[0].Votes) != 1 || s.Timeslots[0].Votes[0].VoterName != "Ali" {
		t.Fatalf("proposer should auto-vote: %+v", s.Timeslots)
	}
}

// Previously, proposing a slot under the name of an existing participant who
// had no password silently cast a vote in their name.
func TestAddTimeslotCannotImpersonate(t *testing.T) {
	e := newEnv(t)
	id := createDynamic(t, e)
	path := "/api/v1/sessions/" + id + "/timeslots"

	expectStatus(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/vote", vote("NoPass", "", "")), 200)
	expectError(t, e.do(t, "POST", path, slot(s10, s11, "NoPass", "", "")), 409, "name_taken_no_password")

	expectStatus(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/vote", vote("Guarded", "pw", "")), 200)
	expectError(t, e.do(t, "POST", path, slot(s10, s11, "Guarded", "", "")), 401, "password_required")
	expectError(t, e.do(t, "POST", path, slot(s10, s11, "Guarded", "nope", "")), 401, "invalid_password")
	expectStatus(t, e.do(t, "POST", path, slot(s10, s11, "Guarded", "pw", "")), 201)

	if n := len(getSession(t, e, id).Timeslots); n != 1 {
		t.Fatalf("rejected proposals must not be stored, got %d slots", n)
	}
}

func TestDeleteTimeslot(t *testing.T) {
	e := newEnv(t)
	id := createDynamic(t, e)
	base := "/api/v1/sessions/" + id + "/timeslots"

	r := e.do(t, "POST", base, slot(s10, s11, "Ali", "pw", ""))
	expectStatus(t, r, 201)
	body := r.JSON(t)
	tsID, token := body["id"].(string), body["token"].(string)

	expectError(t, e.do(t, "DELETE", base+"/"+tsID, nil), 401, "password_required")
	expectError(t, e.do(t, "DELETE", base+"/"+tsID, models.DeleteTimeslotRequest{Password: "bad"}), 401, "invalid_password")
	expectError(t, e.do(t, "DELETE", base+"/nope", models.DeleteTimeslotRequest{Token: token}), 404, "timeslot_not_found")

	// The proposer's own automatic vote doesn't block deletion…
	expectStatus(t, e.do(t, "DELETE", base+"/"+tsID, models.DeleteTimeslotRequest{Token: token}), 200)
	if n := len(getSession(t, e, id).Timeslots); n != 0 {
		t.Fatalf("slot should be gone, have %d", n)
	}

	// …but someone else's vote does.
	r = e.do(t, "POST", base, slot(s11, s12, "Ali", "", token))
	expectStatus(t, r, 201)
	tsID = r.JSON(t)["id"].(string)
	expectStatus(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/vote", vote("Sara", "", "", tsID)), 200)
	expectError(t, e.do(t, "DELETE", base+"/"+tsID, models.DeleteTimeslotRequest{Password: "pw"}), 409, "timeslot_has_votes")
}

func TestDeleteTimeslotFromAnotherSession(t *testing.T) {
	e := newEnv(t)
	a := createDynamic(t, e)
	b := createDynamic(t, e)
	r := e.do(t, "POST", "/api/v1/sessions/"+a+"/timeslots", slot(s10, s11, "", "", ""))
	expectStatus(t, r, 201)
	tsID := r.JSON(t)["id"].(string)
	expectError(t, e.do(t, "DELETE", "/api/v1/sessions/"+b+"/timeslots/"+tsID, nil), 404, "timeslot_not_found")
}
