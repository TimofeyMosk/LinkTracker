package domain

type ErrUserNotExist struct {
}

func (e ErrUserNotExist) Error() string {
	return "user not exists"
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
