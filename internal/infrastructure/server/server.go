package server

import (
	"net/http"
	"time"

	"LinkTracker/internal/application"
	"LinkTracker/internal/infrastructure/clients"
	"LinkTracker/internal/infrastructure/httpapi/links"
	"LinkTracker/internal/infrastructure/httpapi/tgchat"
	"LinkTracker/internal/infrastructure/httpapi/updates"
)

func InitScrapperRouting(scrapper *application.Scrapper) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /links", links.GetLinksHandler{LinkGetter: scrapper})
	mux.Handle("POST /links", links.PostLinkHandler{LinkAdder: scrapper})
	mux.Handle("DELETE /links", links.DeleteLinksHandler{LinkDeleter: scrapper})
	mux.Handle("POST /tg-chat/{id}", tgchat.PostUserHandler{UserAdder: scrapper})
	mux.Handle("DELETE /tg-chat/{id}", tgchat.DeleteUserHandler{UserDeleter: scrapper})

	return mux
}

func InitBotRouting(tgBotAPI *clients.TelegramHTTPClient) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST /updates", updates.PostUpdatesHandler{MessageSender: tgBotAPI})

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
