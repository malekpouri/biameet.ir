package services

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"biameet.ir/models"
)

const (
	maxTitleLen     = 200
	maxNameLen      = 64
	maxNoteLen      = 500
	maxPasswordLen  = 72 // bcrypt rejects longer passwords
	maxCreateSlots  = 50
	maxSessionSlots = 200
	maxSlotLength   = 24 * time.Hour
	maxExpiry       = 366 * 24 * time.Hour

	// Timestamps are stored in the same shape JS's Date.toISOString() produces,
	// so string comparison and ORDER BY match chronological order.
	isoLayout = "2006-01-02T15:04:05.000Z"
)

var hhmm = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func checkText(value, field string, maxLen int, required bool) (string, error) {
	v := strings.TrimSpace(value)
	if required && v == "" {
		return "", invalid(field + " الزامی است")
	}
	if utf8.RuneCountInString(v) > maxLen {
		return "", invalid(field + " بیش از حد طولانی است")
	}
	return v, nil
}

func checkPassword(p string) error {
	if len(p) > maxPasswordLen {
		return invalid("رمز عبور بیش از حد طولانی است")
	}
	return nil
}

func parseISO(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, invalid("فرمت زمان نامعتبر است")
	}
	t = t.UTC()
	if t.Year() < 2000 || t.Year() > 2100 {
		return time.Time{}, invalid("زمان خارج از محدوده است")
	}
	return t, nil
}

// normalizeSlot validates a slot and returns canonical start/end strings.
func normalizeSlot(startS, endS string) (start, end time.Time, err error) {
	if start, err = parseISO(startS); err != nil {
		return
	}
	if end, err = parseISO(endS); err != nil {
		return
	}
	if !end.After(start) {
		err = invalid("زمان شروع باید قبل از زمان پایان باشد")
		return
	}
	if end.Sub(start) > maxSlotLength {
		err = invalid("طول هر زمان حداکثر ۲۴ ساعت است")
	}
	return
}

func validateCreate(req *models.CreateSessionRequest) error {
	var err error
	if req.Title, err = checkText(req.Title, "عنوان جلسه", maxTitleLen, true); err != nil {
		return err
	}
	if req.CreatorName, err = checkText(req.CreatorName, "نام ایجاد کننده", maxNameLen, true); err != nil {
		return err
	}

	if req.Type == "" {
		req.Type = "fixed"
	}
	switch req.Type {
	case "fixed":
		if len(req.Timeslots) == 0 {
			return invalid("حداقل یک زمان برای جلسه لازم است")
		}
		req.DynamicConfig = nil
	case "dynamic", "weekly":
		if err := validateDynamicConfig(req.Type, req.DynamicConfig); err != nil {
			return err
		}
	default:
		return invalid("نوع جلسه نامعتبر است")
	}

	if len(req.Timeslots) > maxCreateSlots {
		return invalid("تعداد زمان‌ها بیش از حد مجاز است")
	}
	for i := range req.Timeslots {
		s, e, err := normalizeSlot(req.Timeslots[i].StartUTC, req.Timeslots[i].EndUTC)
		if err != nil {
			return err
		}
		req.Timeslots[i].StartUTC = s.Format(isoLayout)
		req.Timeslots[i].EndUTC = e.Format(isoLayout)
	}

	if req.ExpiresAtUTC != "" {
		t, err := parseISO(req.ExpiresAtUTC)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if !t.After(now) {
			return invalid("تاریخ انقضا باید در آینده باشد")
		}
		if t.Sub(now) > maxExpiry {
			t = now.Add(maxExpiry)
		}
		req.ExpiresAtUTC = t.Format(isoLayout)
	}
	return nil
}

func validateDynamicConfig(typ string, cfg *models.DynamicConfig) error {
	if cfg == nil {
		return invalid("تنظیمات بازه زمانی الزامی است")
	}
	if !hhmm.MatchString(cfg.MinTime) || !hhmm.MatchString(cfg.MaxTime) {
		return invalid("بازه ساعت نامعتبر است")
	}
	if cfg.MinTime >= cfg.MaxTime {
		return invalid("ساعت شروع بازه باید قبل از ساعت پایان باشد")
	}

	if typ == "dynamic" {
		d, err := parseISO(cfg.DateUTC)
		if err != nil {
			return invalid("تاریخ جلسه الزامی است")
		}
		cfg.DateUTC = d.Format(isoLayout)
		cfg.AllowedDays = nil
		return nil
	}

	// weekly
	cfg.DateUTC = ""
	seen := [7]bool{}
	days := cfg.AllowedDays[:0]
	for _, d := range cfg.AllowedDays {
		if d < 0 || d > 6 {
			return invalid("روز هفته نامعتبر است")
		}
		if !seen[d] {
			seen[d] = true
			days = append(days, d)
		}
	}
	if len(days) == 0 {
		return invalid("حداقل یک روز هفته را انتخاب کنید")
	}
	cfg.AllowedDays = days
	return nil
}

// inDynamicRange loosely checks that a proposed slot falls on the session's
// day. min/max times are in the creator's local zone, which the server doesn't
// know, so the check allows the widest real-world UTC offsets (-12h…+14h).
// The exact HH:MM window is enforced by the browser.
func inDynamicRange(cfg *models.DynamicConfig, start, end time.Time) bool {
	day, err := time.Parse(time.RFC3339Nano, cfg.DateUTC)
	if err != nil {
		return false
	}
	lo := day.Add(-14 * time.Hour)
	hi := day.Add(24*time.Hour + 12*time.Hour)
	return !start.Before(lo) && !end.After(hi)
}
