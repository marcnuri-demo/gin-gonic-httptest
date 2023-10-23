package router

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
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

type context struct {
	router   *gin.Engine
	recorder *httptest.ResponseRecorder
}

func (c *context) beforeEach() {
	c.router = SetupRouter()
	c.recorder = httptest.NewRecorder()
}

func testCase(test func(t *testing.T, c *context)) func(*testing.T) {
	return func(t *testing.T) {
		context := &context{}
		context.beforeEach()
		test(t, context)
	}
}

func TestPostInvalid(t *testing.T) {
	t.Run("Returns 400 status code for empty body", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, httptest.NewRequest("POST", "/", nil))
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"Empty body\"" {
			t.Error("Expected \"Empty body\", got ", c.recorder.Body.String())
		}
	}))
	t.Run("Returns 400 status code for missing Content-Type header", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, httptest.NewRequest("POST", "/", strings.NewReader("{}")))
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"Invalid Content-Type\"" {
			t.Error("Expected \"Invalid Content-Type\", got ", c.recorder.Body.String())
		}
	}))
	t.Run("Returns 400 status code for invalid JSON", testCase(func(t *testing.T, c *context) {
		request := httptest.NewRequest("POST", "/", strings.NewReader("{]"))
		request.Header.Add("Content-Type", "application/json")
		c.router.ServeHTTP(c.recorder, request)
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"Invalid JSON body\"" {
			t.Error("Expected \"Invalid JSON body\", got ", c.recorder.Body.String())
		}
	}))
}

func TestPostValid(t *testing.T) {
	// Given
	reqBuilder := func() *http.Request {
		request := httptest.NewRequest("POST", "/", strings.NewReader(`{
			"name": "test-object",
			"quantity": 1
		}`))
		request.Header.Add("Content-Type", "application/json")
		return request
	}
	t.Run("Returns 201 status code", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, reqBuilder())
		if c.recorder.Code != 201 {
			t.Error("Expected 201, got ", c.recorder.Code)
		}
	}))
	t.Run("Returns saved object with id", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, reqBuilder())
		body := make(map[string]interface{})
		json.Unmarshal(c.recorder.Body.Bytes(), &body)
		matched, _ := regexp.MatchString("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", body["id"].(string))
		if !matched {
			t.Error("Expected object with id, got ", body)
		}
	}))
	t.Run("Returns saved object with provided properties", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, reqBuilder())
		body := make(map[string]interface{})
		json.Unmarshal(c.recorder.Body.Bytes(), &body)
		if body["name"] != "test-object" {
			t.Error("Expected object with name = 'test-object', got ", body)
		}
		if body["quantity"] != 1.0 {
			t.Error("Expected object with quantity = 'test-object', got ", body)
		}
	}))
}
