package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
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
	t.Run("Returns Server header", func(t *testing.T) {
		if recorder.Header().Get("Server") != "gin-gonic/1.33.7" {
			t.Error("Expected 'gin-gonic/1.33.7', got ", recorder.Header().Get("Server"))
		}
	})
}

type context struct {
	router   *gin.Engine
	recorder *httptest.ResponseRecorder
}

func (c *context) beforeEach() {
	entries.Clear()
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
		if c.recorder.Body.String() != "\"empty body\"" {
			t.Error("Expected \"empty body\", got ", c.recorder.Body.String())
		}
	}))
	t.Run("Returns 400 status code for missing Content-Type header", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, httptest.NewRequest("POST", "/", strings.NewReader("{}")))
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"invalid Content-Type\"" {
			t.Error("Expected \"invalid Content-Type\", got ", c.recorder.Body.String())
		}
	}))
	t.Run("Returns 400 status code for invalid JSON", testCase(func(t *testing.T, c *context) {
		request := httptest.NewRequest("POST", "/", strings.NewReader("{]"))
		request.Header.Add("Content-Type", "application/json")
		c.router.ServeHTTP(c.recorder, request)
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"invalid JSON body\"" {
			t.Error("Expected \"invalid JSON body\", got ", c.recorder.Body.String())
		}
	}))
}

func TestPostValid(t *testing.T) {
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
		var body map[string]interface{}
		json.Unmarshal(c.recorder.Body.Bytes(), &body)
		matched, _ := regexp.MatchString("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", body["id"].(string))
		if !matched {
			t.Error("Expected object with id, got ", body)
		}
	}))
	t.Run("Returns saved object with provided properties", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, reqBuilder())
		var body map[string]interface{}
		json.Unmarshal(c.recorder.Body.Bytes(), &body)
		if body["name"] != "test-object" {
			t.Error("Expected object with name = 'test-object', got ", body)
		}
		if body["quantity"] != 1.0 {
			t.Error("Expected object with quantity = 'test-object', got ", body)
		}
	}))
}

func TestGet(t *testing.T) {
	reqBuilder := func() *http.Request {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Add("Accept", "application/json")
		return request
	}
	t.Run("Returns empty list", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, reqBuilder())
		if c.recorder.Code != 200 {
			t.Error("Expected 200, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "[]" {
			t.Error("Expected empty list, got ", c.recorder.Body.String())
		}
	}))
	t.Run("Returns created objects as list", testCase(func(t *testing.T, c *context) {
		// Given
		for i := 1; i <= 3; i++ {
			request := httptest.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(`{
				"object": %d
			}`, i)))
			request.Header.Add("Content-Type", "application/json")
			c.router.ServeHTTP(httptest.NewRecorder(), request)
		}
		// When
		c.router.ServeHTTP(c.recorder, reqBuilder())
		// Then
		var body []map[string]interface{}
		json.Unmarshal(c.recorder.Body.Bytes(), &body)
		if len(body) != 3 {
			t.Error("Expected 3 objects, got ", len(body))
		}
		if !slices.ContainsFunc(body, func(item map[string]interface{}) bool {
			return item["object"] == 1.0
		}) {
			t.Error("Expected object with object = 1, got ", body)
		}
	}))
}

func TestPutInvalid(t *testing.T) {
	t.Run("Returns 400 status code for empty body", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, httptest.NewRequest("PUT", "/static-id", nil))
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"empty body\"" {
			t.Error("Expected \"empty body\", got ", c.recorder.Body.String())
		}
	}))
	t.Run("Returns 400 status code for missing Content-Type header", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, httptest.NewRequest("PUT", "/static-id", strings.NewReader("{}")))
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"invalid Content-Type\"" {
			t.Error("Expected \"invalid Content-Type\", got ", c.recorder.Body.String())
		}
	}))
	t.Run("Returns 400 status code for invalid JSON", testCase(func(t *testing.T, c *context) {
		request := httptest.NewRequest("PUT", "/static-id", strings.NewReader("{]"))
		request.Header.Add("Content-Type", "application/json")
		c.router.ServeHTTP(c.recorder, request)
		if c.recorder.Code != 400 {
			t.Error("Expected 400, got ", c.recorder.Code)
		}
		if c.recorder.Body.String() != "\"invalid JSON body\"" {
			t.Error("Expected \"invalid JSON body\", got ", c.recorder.Body.String())
		}
	}))
}

func TestPutValid(t *testing.T) {
	reqBuilder := func(providedId ...string) (*http.Request, string) {
		var id string
		if len(providedId) > 0 {
			id = providedId[0]
		} else {
			randomId, _ := uuid.NewRandom()
			id = randomId.String()
		}
		request := httptest.NewRequest("PUT", "/"+id, strings.NewReader(`{
			"name": "test-object",
			"quantity": 1
		}`))
		request.Header.Add("Content-Type", "application/json")
		return request, id
	}
	t.Run("Returns 201 status code for insert", testCase(func(t *testing.T, c *context) {
		request, _ := reqBuilder()
		c.router.ServeHTTP(c.recorder, request)
		if c.recorder.Code != 201 {
			t.Error("Expected 201, got ", c.recorder.Code)
		}
	}))
	t.Run("Returns 200 status code for update", testCase(func(t *testing.T, c *context) {
		firstRequest, firstId := reqBuilder()
		c.router.ServeHTTP(httptest.NewRecorder(), firstRequest)
		secondRequest, _ := reqBuilder(firstId)
		c.router.ServeHTTP(c.recorder, secondRequest)
		if c.recorder.Code != 200 {
			t.Error("Expected 200, got ", c.recorder.Code)
		}
	}))
	t.Run("Returns saved object with id", testCase(func(t *testing.T, c *context) {
		request, _ := reqBuilder()
		c.router.ServeHTTP(c.recorder, request)
		var body map[string]interface{}
		json.Unmarshal(c.recorder.Body.Bytes(), &body)
		matched, _ := regexp.MatchString("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", body["id"].(string))
		if !matched {
			t.Error("Expected object with id, got ", body)
		}
	}))
	t.Run("Returns saved object with provided properties", testCase(func(t *testing.T, c *context) {
		request, _ := reqBuilder()
		c.router.ServeHTTP(c.recorder, request)
		var body map[string]interface{}
		json.Unmarshal(c.recorder.Body.Bytes(), &body)
		if body["name"] != "test-object" {
			t.Error("Expected object with name = 'test-object', got ", body)
		}
		if body["quantity"] != 1.0 {
			t.Error("Expected object with quantity = 'test-object', got ", body)
		}
	}))
}

func TestDelete(t *testing.T) {
	t.Run("With existing object returns 204 status code", testCase(func(t *testing.T, c *context) {
		// Given
		request := httptest.NewRequest("POST", "/", strings.NewReader("{}"))
		request.Header.Add("Content-Type", "application/json")
		requestRecorder := httptest.NewRecorder()
		c.router.ServeHTTP(requestRecorder, request)
		var newObject map[string]interface{}
		json.Unmarshal(requestRecorder.Body.Bytes(), &newObject)
		// When
		c.router.ServeHTTP(c.recorder, httptest.NewRequest("DELETE", "/"+newObject["id"].(string), nil))
		// Then
		if c.recorder.Code != 204 {
			t.Error("Expected 204, got ", c.recorder.Code)
		}
	}))
	t.Run("With missing object returns 404 status code", testCase(func(t *testing.T, c *context) {
		c.router.ServeHTTP(c.recorder, httptest.NewRequest("DELETE", "/a-random-id", nil))
		if c.recorder.Code != 404 {
			t.Error("Expected 404, got ", c.recorder.Code)
		}
	}))
}
