package router

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/orcaman/concurrent-map/v2"
	"strings"
)

var entries = cmap.New[map[string]interface{}]()

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/", addCommonHeaders, get, fallbackGet)
	router.POST("/", addCommonHeaders, post)
	router.PUT("/:id", addCommonHeaders, put)
	router.DELETE("/:id", addCommonHeaders, remove)
	return router
}

func containsHeader(c *gin.Context) func(key string, value string) bool {
	return func(key string, value string) bool {
		values := c.GetHeader(key)
		for _, v := range strings.Split(values, ",") {
			if strings.TrimSpace(v) == value {
				return true
			}
		}
		return false
	}
}

func addCommonHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store")
	c.Header("Server", "gin-gonic/1.33.7")
}

func jsonRequestBodyBody(c *gin.Context) (map[string]interface{}, error) {
	if c.Request.ContentLength == 0 {
		return nil, errors.New("empty body")
	}
	if !containsHeader(c)("Content-Type", "application/json") {
		return nil, errors.New("invalid Content-Type")
	}
	data := make(map[string]interface{})
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		err = errors.New("invalid JSON body")
	}
	return data, err
}

func fallbackGet(c *gin.Context) {
	if c.Writer.Size() > -1 {
		return
	}
	c.IndentedJSON(200, "Cocktail service")
}

func get(c *gin.Context) {
	if !containsHeader(c)("Accept", "application/json") {
		return
	}
	r := make([]map[string]interface{}, 0, entries.Count())
	for item := range entries.IterBuffered() {
		r = append(r, item.Val)
	}
	c.IndentedJSON(200, r)
	return
}

func post(c *gin.Context) {
	data, err := jsonRequestBodyBody(c)
	if err != nil {
		c.IndentedJSON(400, err.Error())
		return
	}
	id, _ := uuid.NewRandom()
	data["id"] = id.String()
	entries.Set(id.String(), data)
	c.IndentedJSON(201, data)
}

func put(c *gin.Context) {
	id := c.Param("id")
	data, err := jsonRequestBodyBody(c)
	if err != nil {
		c.IndentedJSON(400, err.Error())
		return
	}
	data["id"] = id
	var status int
	entries.Upsert(id, data, func(exists bool, valueInMap map[string]interface{}, newValue map[string]interface{}) map[string]interface{} {
		if exists {
			status = 200
		} else {
			status = 201
		}
		return newValue
	})
	c.IndentedJSON(status, data)
}

func remove(c *gin.Context) {
	id := c.Param("id")
	if !entries.Has(id) {
		c.IndentedJSON(404, "Not found")
		return
	}
	entries.Remove(id)
	c.Status(204)
}
