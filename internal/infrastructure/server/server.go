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

func InitScrapperRouting(s *scrapper.Scrapper) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /links", links.GetLinksHandler{LinkGetter: s})
	mux.Handle("POST /links", links.PostLinksHandler{LinkAdder: s})
	mux.Handle("DELETE /links", links.DeleteLinksHandler{LinkDeleter: s})
	mux.Handle("PUT /links", links.PutLinksHandler{LinkUpdater: s})

	mux.Handle("POST /tg-chat/{id}", tgchat.PostUserHandler{UserAdder: s})
	mux.Handle("DELETE /tg-chat/{id}", tgchat.DeleteUserHandler{UserDeleter: s})

	mux.Handle("POST /states", states.PostStatesHandler{StateCreator: s})
	mux.Handle("DELETE /states", states.DeleteStatesHandler{StateDeleter: s})
	mux.Handle("PUT /states", states.PutStatesHandler{StateUpdater: s})
	mux.Handle("GET /states", states.GetStatesHandler{StateGetter: s})

	return mux
}

func InitBotRouting(b *bot.Bot) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST /updates", updates.PostUpdatesHandler{UpdateSender: b})

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
