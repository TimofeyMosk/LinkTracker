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
	mux.Handle("POST "+"/links", handlers.PostLinkHandler{Scrapper: scrapper})
	mux.Handle("DELETE "+"/links", handlers.DeleteLinksHandler{Scrapper: scrapper})
	mux.Handle("POST "+"/tg-chat/{id}", handlers.PostUserHandler{Scrapper: scrapper})
	mux.Handle("DELETE "+"/tg-chat/{id}", handlers.DeleteUserHandler{Srapper: scrapper})

	return mux
}

func InitBotRouting(bot *application.Bot) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST "+"/updates", handlers.PostUpdatesHandler{Bot: bot})

	return mux
}

func InitServer(addr string, handler http.Handler, readTimeout, writeTimeout time.Duration) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}
