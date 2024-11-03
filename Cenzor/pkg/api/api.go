package api

import (
	"Cenzor/pkg/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	reqIDStr string = "requset_id"
)

type API struct {
	r *mux.Router
}

func New() *API {
	api := API{}
	api.r = mux.NewRouter()
	api.endpoints()
	return &api
}

func (api *API) endpoints() {
	api.r.Use(logMiddleware)
	api.r.HandleFunc("/cenzor", api.cenzor).Methods(http.MethodPost, http.MethodOptions)
}

func (api *API) Router() *mux.Router {
	return api.r
}

// logMiddleware logs the request ID, IP, method, URL and duration of the request.
// It is meant to be used as a middleware for the API.
// It logs the request duration using the time.Since function.
// The request ID is taken from the request URL query parameter "requset_id".
// The IP is taken from the request RemoteAddr.
// The method is taken from the request Method.
// The URL is taken from the request URL.
// The duration is taken from the time.Since function.
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(begin time.Time) {
			requestID := r.URL.Query().Get(reqIDStr)
			ip := r.RemoteAddr
			log.Printf("RequestID: %s IP: %s Method: %s URL: %s Duration: %s", requestID, ip, r.Method, r.URL, time.Since(begin))
		}(time.Now())
		next.ServeHTTP(w, r)
	})
}

// cenzor censors a comment and returns a 200 OK status if the comment is clean,
// and a 400 Bad Request status if the comment is censored.
// The comment is passed in the request body as a JSON object.
// The request ID is passed as a query parameter "requset_id".
// The IP is passed as a query parameter "ip".
// The request duration is logged using the logMiddleware.
func (api *API) cenzor(w http.ResponseWriter, r *http.Request) {
	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	status := censored(comment.Content)

	if status {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// censored checks if a comment contains swear words.
// It takes a string as a parameter and returns a boolean.
// If the comment contains any of the swear words, it returns false.
// Otherwise, it returns true.
func censored(comment string) bool {
	swearWords := []string{"йцукен", "ячсмит", "пролсд"}

	for _, swearWord := range swearWords {
		if strings.Contains(strings.ToLower(comment), swearWord) {
			return false
		}
	}
	return true
}
