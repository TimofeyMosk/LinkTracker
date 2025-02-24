package HTTPscrapper

type errUserNotExist struct {
	msg string
}

func (e errUserNotExist) Error() string {
	e.msg = "user not exist"
	return e.msg
}
