package api

/*
Form errors
 */
type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(msg string, code int) Error {
	return Error{msg, code}
}

var (
	ErrInvalidEmail     = NewError("email is not valid", 100)
	ErrPasswordTooShort = NewError("password must be at least 6 characters long", 101)
	ErrEntityNameTooShort = NewError("entity name must be at least 3 characters long", 101)
)
