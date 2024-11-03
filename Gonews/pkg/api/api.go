package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/suxrobshukurov/gonews/pkg/paginate"
	"github.com/suxrobshukurov/gonews/pkg/storage"
)

const (
	reqIDStr string = "requset_id"
)

// API struct
type API struct {
	r  *mux.Router
	db storage.Interface
}

// New creates a new API
func New(db storage.Interface) *API {
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

	api.r.HandleFunc("/news", api.postsHandler).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/id", api.postById).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/filter", api.filternews).Methods(http.MethodGet, http.MethodOptions)
	// web app
	// api.r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// headersMiddleware sets the Content-Type to application/json and Access-Control-Allow-Origin to *.
// It also catches and handles OPTIONS requests by returning a 204 No Content status.
func headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

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


// postById handles the HTTP GET request to retrieve a specific post by its ID.
// It extracts the "id" query parameter from the request URL, converts it to an integer,
// and attempts to fetch the corresponding post from the database using the PostByID method.
// If the ID is invalid or if the post cannot be found, it returns an appropriate HTTP error response.
// If the post is found, it encodes the post into JSON and writes it to the response.
// Handles errors for invalid ID format, database retrieval issues, and JSON encoding failures.
func (api *API) postById(w http.ResponseWriter, r *http.Request) {
	strID := r.URL.Query().Get("id")

	id, err := strconv.Atoi(strID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Invalid post ID. Error: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	post, err := api.db.PostByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Post not found"}`))
			return
		}
		http.Error(w, fmt.Sprintf(`{"error": "Can't get post by ID. Error: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't encode post. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}


// postsHandler handles the HTTP GET request to retrieve a paginated list of posts.
// It first fetches the total count of posts from the database.
// If an error occurs while fetching the count, it returns an HTTP 500 error response.
// If successful, it invokes the handlePagination method to manage pagination
// and fetches posts using the Posts method, encoding the result in JSON format.
func (api *API) postsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := api.db.Count()
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't get count. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	api.handlePagination(w, r, count, func(start, limit int) ([]storage.Post, error) {
		return api.db.Posts(start, limit)
	})
}

// filternews handles the HTTP GET request to retrieve a paginated list of posts
// that match a search pattern.
//
// It first fetches the total count of posts matching the search pattern from the database.
// If an error occurs while fetching the count, it returns an HTTP 500 error response.
// If successful, it invokes the handlePagination method to manage pagination
// and fetches posts using the Filter method, encoding the result in JSON format.
func (api *API) filternews(w http.ResponseWriter, r *http.Request) {
	strSearch := r.URL.Query().Get("s")
	count, err := api.db.CountOfFilter(strSearch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't get count. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	api.handlePagination(w, r, count, func(start, limit int) ([]storage.Post, error) {
		return api.db.Filter(strSearch, start, limit)
	})

}

// handlePagination handles the pagination of posts.
// It takes a total count of posts, a HTTP request, and a function to fetch posts.
// It extracts the "page" query parameter from the request URL, and uses it to calculate the start index to fetch posts.
// If the total count is 0, it returns an empty paginated response.
// If the total count is greater than 0, it fetches posts using the provided function and calculates the total number of pages.
// It then encodes the paginated response into JSON and writes it to the response.
// It handles errors for invalid page number, database retrieval issues, and JSON encoding failures.
func (api *API) handlePagination(w http.ResponseWriter, r *http.Request, totalCount int, fetchPosts func(start, limit int) ([]storage.Post, error)) {
	strPage := r.URL.Query().Get("page")
	page, err := strconv.Atoi(strPage)
	if err != nil || page < 1 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	if totalCount == 0 {
		res := paginate.Paginate{
			Posts:      []storage.Post{},
			Pagination: paginate.Pagination{CurrentPage: 1, TotalPages: 1, NumberOfPosts: 0},
		}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, fmt.Sprintf("Can't encode response. Error: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	start := paginate.PostsPerPage * (page - 1)
	posts, err := fetchPosts(start, paginate.PostsPerPage)
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't get posts. Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	pages := (totalCount / paginate.PostsPerPage) + 1
	var numberOfPosts int
	if page < pages {
		numberOfPosts = len(posts)
	} else {
		numberOfPosts = totalCount % paginate.PostsPerPage
	}

	p := paginate.Pagination{
		CurrentPage:   page,
		TotalPages:    pages,
		NumberOfPosts: numberOfPosts,
	}
	res := paginate.Paginate{
		Posts:      posts,
		Pagination: p,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, fmt.Sprintf("Can't encode posts. Error: %s", err.Error()), http.StatusInternalServerError)
	}

}
