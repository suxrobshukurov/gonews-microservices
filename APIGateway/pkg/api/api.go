package api

import (
	"APIGateway/pkg/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"math/rand"
	"time"

	"github.com/gorilla/mux"
)

const (
	reqIDStr    string = "requset_id"
	newsUrl     string = "http://localhost:8081/news"
	commentsUrl string = "http://localhost:8082/comments"
	cenzorUrl   string = "http://localhost:8083/cenzor"
)

type API struct {
	r *mux.Router
}

func New() *API {
	api := &API{}
	api.r = mux.NewRouter()
	api.endpoints()
	return api
}

func (api *API) endpoints() {
	api.r.Use(requestIDMiddleware)
	api.r.Use(headersMiddleware)
	api.r.Use(logMiddleware)
	api.r.HandleFunc("/news", api.news).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/filter", api.newsFilter).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/id", api.detailedNews).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/comment", api.addComment).Methods(http.MethodPost, http.MethodOptions)
}
func (api *API) Router() *mux.Router {
	return api.r
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

// requestIDMiddleware adds a request ID to the request context if it doesn't already exist.
// If the request ID already exists in the URL query parameters, it is used.
// Otherwise, a new request ID is generated using the idGenerator function.
// The request ID is stored in the request context under the key reqIDStr.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.URL.Query().Get(reqIDStr)
		if requestID == "" {
			requestID = idGenerator()
		}
		ctx := context.WithValue(context.Background(), reqIDStr, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
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
			requestID := r.Context().Value(reqIDStr).(string)
			ip := r.RemoteAddr
			log.Printf("RequestID: %s IP: %s Method: %s URL: %s Duration: %s", requestID, ip, r.Method, r.URL, time.Since(begin))
		}(time.Now())
		next.ServeHTTP(w, r)
	})
}

// news returns a list of news items in JSON format.
// The page parameter is required and specifies the page number of the list of news items to return.
// The request ID is taken from the request context.
// If the request ID doesn't exist in the request context, a new one is generated.
// The request ID is included in the URL query parameter "requset_id".
// The function returns a 200 OK status with the list of news items in JSON format.
// If there is an error, the function returns a 400 Bad Request status with the error message.
func (api *API) news(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(reqIDStr).(string)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	urlStr := newsUrl + "?page=" + page + "&" + reqIDStr + "=" + reqID

	resp, err := http.Get(urlStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}

// newsFilter returns a list of news items in JSON format that match the search query.
// The page parameter is required and specifies the page number of the list of news items to return.
// The request ID is taken from the request context.
// If the request ID doesn't exist in the request context, a new one is generated.
// The request ID is included in the URL query parameter "requset_id".
// The function returns a 200 OK status with the list of news items in JSON format.
// If there is an error, the function returns a 400 Bad Request status with the error message.
func (api *API) newsFilter(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(reqIDStr).(string)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	strSearch := r.URL.Query().Get("s")
	urlStr := newsUrl + "/filter?s=" + strSearch + "&page=" + page + "&" + reqIDStr + "=" + reqID

	resp, err := http.Get(urlStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// detailedNews returns a detailed news item in JSON format.
// The post ID is required and is passed as a query parameter "id".
// The request ID is taken from the request context.
// If the request ID doesn't exist in the request context, a new one is generated.
// The request ID is included in the URL query parameter "requset_id".
// The function returns a 200 OK status with the detailed news item in JSON format.
// If there is an error, the function returns a 400 Bad Request status with the error message.
func (api *API) detailedNews(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(reqIDStr).(string)
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "ID not found.", http.StatusBadRequest)
		return
	}

	result := make(chan interface{}, 2)
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		post, err := getPost(postID, reqID)
		if err != nil {
			result <- err
			return
		}
		result <- post
	}()

	go func() {
		defer wg.Done()
		comments, err := getComments(postID, reqID)
		if err != nil {
			result <- err
			return
		}
		result <- comments
	}()

	go func() {
		wg.Wait()
		close(result)
	}()

	var post models.PostFullDetailed
	var comments []models.Comment

	for res := range result {
		switch res := res.(type) {
		case models.PostFullDetailed:
			post = res
		case []models.Comment:
			comments = res
		case error:
			http.Error(w, res.Error(), http.StatusBadRequest)
			return
		}
	}

	post.Comments = comments
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// addComment adds a new comment to the database.
// The request body should contain a valid Comment struct.
// The comment is first sent to the cenzor service for moderation.
// If the cenzor service returns 200 OK, the comment is added to the database.
// If there is an error during the cenzor request, it returns a 400 Bad Request status.
// If the cenzor service doesn't return 200 OK, the comment is not added to the database.
// If there is an error during the database request, it returns a 500 Internal Server Error status.
// Otherwise, it returns a 200 OK status.
func (api *API) addComment(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(reqIDStr).(string)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	cenzorReq, err := http.NewRequest(
		http.MethodPost,
		cenzorUrl+"?"+reqIDStr+"="+reqID,
		bytes.NewBuffer(body),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cenzorReq.Header.Set("Content-Type", "application/json")
	cenzorClient := &http.Client{}
	cenzorResponse, err := cenzorClient.Do(cenzorReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer cenzorResponse.Body.Close()

	if cenzorResponse.StatusCode == http.StatusOK {
		req, err := http.NewRequest(
			http.MethodPost,
			commentsUrl+"?"+reqIDStr+"="+reqID,
			bytes.NewBuffer(body),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusOK {
			w.WriteHeader(http.StatusOK)
		} else {
			// w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Failed to save comment", http.StatusBadRequest)
		}
	} else {
		// w.WriteHeader(http.StatusInternalServerError)
		// w.Write([]byte("The comment was not censored"))
		http.Error(w, "Comment was not censored", http.StatusBadRequest)

	}
}

// getPost retrieves a post by its id from the news service
// and returns a models.PostFullDetailed struct. If there is an error
// during the request, it returns the error. If the news service returns
// a non-OK status, it returns an error with the status text as the message.
// Otherwise, it unmarshals the JSON response into a models.PostFullDetailed
// and returns it.
func getPost(id string, reqID string) (models.PostFullDetailed, error) {
	urlStr := newsUrl + "/id?id=" + id + "&" + reqIDStr + "=" + reqID
	resp, err := http.Get(urlStr)
	if err != nil {
		return models.PostFullDetailed{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.PostFullDetailed{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return models.PostFullDetailed{}, errors.New(string(body))
	}

	var post models.PostFullDetailed
	if err := json.Unmarshal(body, &post); err != nil {
		return models.PostFullDetailed{}, err
	}

	return post, nil

}

// getComments retrieves a list of comments for a post from the comments service
// and returns the list of comments and an error if any. If there is an error
// during the request, it returns the error. If the comments service returns a
// non-OK status, it returns an error with the status text as the message.
// Otherwise, it unmarshals the JSON response into a list of models.Comment
// and returns it.
func getComments(id string, reqID string) ([]models.Comment, error) {
	urlStr := commentsUrl + "?id_post=" + id + "&" + reqIDStr + "=" + reqID
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var comments []models.Comment
	if err := json.Unmarshal(body, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

// idGenerator returns a random 7-digit number as a string. It is used as a request ID
// for tracing requests through the system. It is not guaranteed to be unique, but
// the probability of a collision is very low.
func idGenerator() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.Itoa(r.Intn(10_000_000) + 100_000)
}
