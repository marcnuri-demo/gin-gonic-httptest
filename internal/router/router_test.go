package router

import (
	"net/http/httptest"
	"testing"
)

func TestFallbackGet(t *testing.T) {
	router := SetupRouter()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))
	t.Run("Returns 200 status code", func(t *testing.T) {
		if recorder.Code != 200 {
			t.Error("Expected 200, got ", recorder.Code)
		}
	})
	t.Run("Returns app name", func(t *testing.T) {
		if recorder.Body.String() != "\"Cocktail service\"" {
			t.Error("Expected '\"Cocktail service\"', got ", recorder.Body.String())
		}
	})
}
