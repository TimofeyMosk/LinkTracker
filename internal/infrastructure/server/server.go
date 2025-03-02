package server

import (
	"net/http"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers"
)

func InitScrapperRouting(scrapper *application.Scrapper) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET "+"/links", handlers.GetLinksHandler{Scrapper: scrapper})
	mux.Handle("POST "+"/links", handlers.PostLinksHandler{Scrapper: scrapper})
	mux.Handle("DELETE "+"/links", handlers.DeleteLinksHandler{Scrapper: scrapper})
	mux.Handle("POST "+"/tg-chat/{id}", handlers.RegisterUserHandler{Scrapper: scrapper})
	mux.Handle("DELETE "+"/tg-chat/{id}", handlers.DeleteUserHandler{Srapper: scrapper})

	return mux
}

func InitBotRouting() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST"+"/updates", nil)

	return mux
}

func InitServer(addr string, handler http.Handler, readTimeout, writeTumeout time.Duration) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTumeout,
	}
}
