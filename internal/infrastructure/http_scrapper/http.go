package infrastructure

import (
	"fmt"
	"net/http"
	"strconv"
)

type UserHandler struct{}

func (UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Hello from UserHandler\n"))
}

type UserRegistrationHandler struct{}

func (UserRegistrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := strconv.Atoi(r.PathValue("id")); err == nil && r.PathValue("id") != "" {

	}
	_, _ = fmt.Fprintf(w, "User id=%s, registrated \n", r.PathValue("id"))
}

type UserDeleteHandler struct{}

func (UserDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "User id= %s, deleted \n", r.PathValue("id"))
}

func InitRouting() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET "+"/links", UserHandler{})
	mux.Handle("POST "+"/links", UserHandler{})
	mux.Handle("DELETE "+"/links", UserHandler{})
	mux.Handle("POST "+"/tg-chat/{id}", UserRegistrationHandler{})
	mux.Handle("DELETE "+"/tg-chat/{id}", UserDeleteHandler{})
	return mux
}

func InitServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
