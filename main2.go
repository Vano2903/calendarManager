package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	handler *HandlerHttp
)

func init() {
	var err error
	if handler, err = NewHandlerHttp(googleCredentialsFile); err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.Use(handler.loggerMiddleware)

	//serve static files
	r.PathPrefix("/static").Handler(http.FileServer(http.Dir("./static/")))
	//!need to know the pages first
	// r.HandleFunc("/oauth", oauthGooglePageHandler).Methods("GET")

	sheets := r.PathPrefix("/sheets").Subrouter()
	sheets.Use(handler.AccessTokenMiddleware)
	sheets.HandleFunc("/new", handler.CreateSheetHandler).Methods("GET")
	sheets.HandleFunc("/events", handler.GetEventsHandler).Methods("GET")

	calendar := r.PathPrefix("/calendar").Subrouter()
	calendar.Use(handler.AccessTokenMiddleware)
	calendar.HandleFunc("/update", handler.UpdateCalendarHandler).Methods("GET")

	oauth := r.PathPrefix("/auth").Subrouter()
	oauth.HandleFunc("/google/login", handler.OauthGoogleHandler).Methods("GET")
	oauth.HandleFunc("/google/callback", handler.OauthGoogleCallbackHandler).Methods("GET")

	fmt.Println("listening on port 8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}
}
