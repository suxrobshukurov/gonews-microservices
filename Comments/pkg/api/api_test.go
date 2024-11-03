package api

import (
	"Comments/pkg/db"
	"Comments/pkg/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupAPI(t *testing.T) *API {
	os.Setenv("connstr", "postgres://test_user:password@localhost:5432/test_comments?sslmode=disable")
	mockDB, err := db.New()
	if err != nil {
		t.Fatal(err)
	}
	api := New(mockDB)
	return api
}

func TestAddComment(t *testing.T) {
	api := setupAPI(t)

	newComment := models.Comment{
		PostID:   1,
		ParentID: 0,
		Content:  "test comment",
		AddTime:  time.Now().Unix(),
	}

	reqBody, _ := json.Marshal(newComment)
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]int
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotZero(t, resp["id"])
}

func TestUpdateComment(t *testing.T) {
	api := setupAPI(t)

	updatedComment := models.Comment{
		ID:       1,
		PostID:   1,
		ParentID: 0,
		Content:  "updated comment",
		AddTime:  time.Now().Unix(),
	}

	reqBody, _ := json.Marshal(updatedComment)
	req := httptest.NewRequest(http.MethodPut, "/comments/1", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteComment(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodDelete, "/comments/1", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAddCommentInvalidData(t *testing.T) {
	api := setupAPI(t)

	invalidData := []byte(`{"PostID": "Invalid_id"}`)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(invalidData))
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCommentsInvalidID(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/comments?id_post=invalid_id", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteCommentInvalidID(t *testing.T) {
	api := setupAPI(t)

	req := httptest.NewRequest(http.MethodDelete, "/comments/invalid_id", nil)
	w := httptest.NewRecorder()

	api.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
