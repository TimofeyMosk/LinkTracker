package domain

import "fmt"

type ErrUserNotExist struct {
}

func (e ErrUserNotExist) Error() string {
	return "user not exists"
}

type ErrUserAlreadyExist struct{}

func (e ErrUserAlreadyExist) Error() string {
	return "user already exists"
}

type ErrEmptyString struct{}

func (e ErrEmptyString) Error() string {
	return "empty string"
}

type ErrQuestionNotFound struct {
}

func (e ErrQuestionNotFound) Error() string {
	return "question not found"
}

type ErrWrongURL struct {
}

func (e ErrWrongURL) Error() string {
	return "wrong link, expected host ‘stackoverflow.com’ or ‘github.com’"
}

type ErrUnsupportedHost struct{}

func (e ErrUnsupportedHost) Error() string {
	return "unsupported host"
}

type ErrRegistrationUser struct{}

func (e ErrRegistrationUser) Error() string {
	return "failed to register a user"
}

type ErrDeletionUser struct{}

func (e ErrDeletionUser) Error() string {
	return "failed to deletion a user"
}

type ErrAPI struct {
	Code             string
	Description      string
	ExceptionMessage string
	ExceptionName    string
	Stacktrace       []string
}

func (e ErrAPI) Error() string { //nolint:gocritic // This is an error, it should not address by pointer
	return fmt.Sprintf("api error %s: %s", e.Code, e.ExceptionMessage)
}

type ErrUnexpectedStatusCode struct {
	StatusCode int
}

func (e ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("unexpected status code [%d]", e.StatusCode)
}

type ErrNoRequiredAttribute struct {
	Attribute string
}

func (e ErrNoRequiredAttribute) Error() string {
	return fmt.Sprintf("no required attribute [%s]", e.Attribute)
}

type ErrLinkAlreadyTracking struct{}

func (e ErrLinkAlreadyTracking) Error() string {
	return "link already tracking"
}

type ErrUpdatesNotFound struct{}

func (e ErrUpdatesNotFound) Error() string {
	return "updates not found"
}
