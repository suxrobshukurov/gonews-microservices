package api

import (
	"Cenzor/pkg/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCenzor(t *testing.T) {
	api := New()

	tests := []struct {
		name         string
		comment      models.Comment
		expectedCode int
	}{
		{
			name: "Valid comment",
			comment: models.Comment{
				Content: "This is a clean comment.",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid comment with swear word",
			comment: models.Comment{
				Content: "This comment contains йцукен.",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Another invalid comment",
			comment: models.Comment{
				Content: "This is another bad comment: ячсмит.",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Empty comment",
			comment: models.Comment{
				Content: "",
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.comment)
			if err != nil {
				t.Fatalf("Failed to marshal comment: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/cenzor", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			api.Router().ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}
		})
	}
}
