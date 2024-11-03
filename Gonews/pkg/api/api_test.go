package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suxrobshukurov/gonews/pkg/paginate"
	"github.com/suxrobshukurov/gonews/pkg/storage"
	"github.com/suxrobshukurov/gonews/pkg/storage/postgres"
)

func setupAPI(t *testing.T) *API {
	os.Setenv("connstr", "postgres://test_user:password@localhost:5432/for_testing?sslmode=disable")
	db, err := postgres.New()
	if err != nil {
		t.Fatal(err)
	}

	return New(db)
}

func TestGetPosts(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/news?page=1", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response paginate.Paginate
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	assert.NotEmpty(t, response.Posts, "Posts should not be empty")
	assert.Equal(t, 1, response.Pagination.CurrentPage, "Current page should be 1")
}

func TestGetPostByID(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/news/id?id=1", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var post storage.Post
	err := json.Unmarshal(w.Body.Bytes(), &post)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	assert.Equal(t, 1, post.ID, "Post ID should be 1")
	assert.Equal(t, "Test Post 1", post.Title, "Post title should match")
}

func TestAPI_addPostsHandler(t *testing.T) {
	api := setupAPI(t)

	post := storage.Post{
		ID:      6,
		Title:   "Test Post",
		Content: "This is a test content",
		PubTime: 20241100,
		Link:    "http://example.com/test-post",
	}
	posts := []storage.Post{post}

	body, _ := json.Marshal(posts)
	req, err := http.NewRequest(http.MethodPost, "/news/add", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	api.Router().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestGetPostByInvalidID(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/news/id?id=invalid_id", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFilterPosts(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/news/filter?page=1&s=Test", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response paginate.Paginate
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	assert.NotEmpty(t, response.Posts, "Filtered posts should not be empty")
	assert.Equal(t, 1, response.Pagination.CurrentPage, "Current page should be 1")
}

func TestFilterPostsInvalidPage(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/news/filter?page=invalid_id&s=Test", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
