package server

import (
	"net/http"
	"time"

	"LinkTracker/internal/application/bot"

	"LinkTracker/internal/application/scrapper"
	"LinkTracker/internal/infrastructure/httpapi/links"
	"LinkTracker/internal/infrastructure/httpapi/states"
	"LinkTracker/internal/infrastructure/httpapi/tgchat"
	"LinkTracker/internal/infrastructure/httpapi/updates"
)

func InitScrapperRouting(scrapper *scrapper.Scrapper) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /links", links.GetLinksHandler{LinkGetter: scrapper})
	mux.Handle("POST /links", links.PostLinksHandler{LinkAdder: scrapper})
	mux.Handle("DELETE /links", links.DeleteLinksHandler{LinkDeleter: scrapper})
	mux.Handle("PUT /links", links.PutLinksHandler{LinkUpdater: scrapper})

	mux.Handle("POST /tg-chat/{id}", tgchat.PostUserHandler{UserAdder: scrapper})
	mux.Handle("DELETE /tg-chat/{id}", tgchat.DeleteUserHandler{UserDeleter: scrapper})

	mux.Handle("POST /states", states.PostStatesHandler{StateCreator: scrapper})
	mux.Handle("DELETE /states", states.DeleteStatesHandler{StateDeleter: scrapper})
	mux.Handle("PUT /states", states.PutStatesHandler{StateUpdater: scrapper})
	mux.Handle("GET /states", states.GetStatesHandler{StateGetter: scrapper})

	return mux
}

func InitBotRouting(bot *bot.Bot) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST /updates", updates.PostUpdatesHandler{UpdateSender: bot})

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
