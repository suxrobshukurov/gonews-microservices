package api

import (
	"Comments/pkg/db"
	"Comments/pkg/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const (
	reqIDStr string = "requset_id"
)

// API is the API struct
type API struct {
	r  *mux.Router
	db *db.DB
}

// New creates a new API
func New(db *db.DB) *API {
	api := API{}
	api.db = db
	api.r = mux.NewRouter()
	api.endpoints()
	return &api
}

// Router returns the router to use
// as the HTTP server argument.
func (api *API) Router() *mux.Router {
	return api.r
}

// Registration of API methods in the request router.
func (api *API) endpoints() {
	api.r.Use(headersMiddleware)
	api.r.Use(logMiddleware)
	api.r.HandleFunc("/comments", api.comments).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/comments", api.addComment).Methods(http.MethodPost, http.MethodOptions)
	api.r.HandleFunc("/comments/{id}", api.updateComment).Methods(http.MethodPut, http.MethodOptions)
	api.r.HandleFunc("/comments/{id}", api.deleteComment).Methods(http.MethodDelete, http.MethodOptions)
}

// headersMiddleware sets the Content-Type to application/json and Access-Control-Allow-Origin to *.
// It also sets the Access-Control-Allow-Methods to GET, POST, PUT, DELETE, OPTIONS and
// Access-Control-Allow-Headers to Content-Type, Authorization.
// It also catches and handles OPTIONS requests by returning a 204 No Content status.
func headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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


// addComment adds a new comment to the database.
// The request body should contain a valid Comment struct.
// If the request body is invalid, it returns a 400 Bad Request status.
// If there is an error when adding the comment, it returns a 500 Internal Server Error status.
// Otherwise, it returns a 200 OK status and a JSON response containing the ID of the newly added comment.
func (api *API) addComment(w http.ResponseWriter, r *http.Request) {
	var c models.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	// id, err := api.db.AddComment(c)
	_, err := api.db.AddComment(c)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add comment. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(map[string]int{"id": id})
}


// updateComment updates an existing comment in the database.
// The request body should contain a valid Comment struct.
// If the request body is invalid, it returns a 400 Bad Request status.
// If there is an error when updating the comment, it returns a 500 Internal Server Error status.
// Otherwise, it returns a 204 No Content status.
func (api *API) updateComment(w http.ResponseWriter, r *http.Request) {
	var c models.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if err := api.db.UpdateComment(c); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update comment. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}


// deleteComment deletes a comment from the database.
// The comment ID is passed as a URL parameter "id".
// If the comment ID is invalid, it returns a 400 Bad Request status.
// If there is an error when deleting the comment, it returns a 500 Internal Server Error status.
// Otherwise, it returns a 204 No Content status.
func (api *API) deleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid comment ID. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if err := api.db.DeleteComment(id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete comment. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}


// comments returns a list of comments for the given post ID.
// The post ID is passed as a query parameter "id_post".
// If the post ID is invalid, it returns a 400 Bad Request status.
// If there is an error when retrieving the comments, it returns a 500 Internal Server Error status.
// Otherwise, it returns a 200 OK status and a JSON response containing the list of comments.
func (api *API) comments(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_post")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid post ID. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	comments, err := api.db.Comments(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get comments. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode comments. Error: %s", err.Error()), http.StatusInternalServerError)
	}
}
