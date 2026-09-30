package services

import "net/http"

// Error is a client-facing failure. Code is a stable machine-readable string
// the frontend matches on; Message is shown to the user.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Code }

func newErr(status int, code, msg string) *Error {
	return &Error{Status: status, Code: code, Message: msg}
}

var (
	ErrSessionNotFound  = newErr(http.StatusNotFound, "session_not_found", "جلسه مورد نظر یافت نشد")
	ErrTimeslotNotFound = newErr(http.StatusNotFound, "timeslot_not_found", "زمان مورد نظر یافت نشد")
	ErrSessionExpired   = newErr(http.StatusGone, "session_expired", "مهلت این جلسه به پایان رسیده است")

	ErrPasswordRequired    = newErr(http.StatusUnauthorized, "password_required", "برای این کار وارد کردن رمز عبور الزامی است")
	ErrInvalidPassword     = newErr(http.StatusUnauthorized, "invalid_password", "رمز عبور اشتباه است")
	ErrNameTakenNoPassword = newErr(http.StatusConflict, "name_taken_no_password", "این نام قبلاً بدون رمز عبور ثبت شده و امکان ویرایش آن وجود ندارد")
	ErrConcurrentUpdate    = newErr(http.StatusConflict, "concurrent_update", "اطلاعات هم‌زمان تغییر کرد، لطفاً دوباره تلاش کنید")

	ErrDuplicateTimeslot = newErr(http.StatusConflict, "duplicate_timeslot", "این زمان قبلاً ثبت شده است")
	ErrTimeslotHasVotes  = newErr(http.StatusConflict, "timeslot_has_votes", "زمانی که دیگران به آن رای داده‌اند قابل حذف نیست")
	ErrTooManyTimeslots  = newErr(http.StatusConflict, "too_many_timeslots", "تعداد زمان‌های این جلسه به حداکثر رسیده است")
	ErrFixedSession      = newErr(http.StatusBadRequest, "fixed_session", "در این جلسه امکان افزودن زمان جدید وجود ندارد")
	ErrOutOfRange        = newErr(http.StatusBadRequest, "out_of_range", "زمان انتخابی خارج از بازه مجاز جلسه است")
)

// invalid builds a 400 validation error with a user-facing message.
func invalid(msg string) *Error {
	return newErr(http.StatusBadRequest, "invalid_input", msg)
}
