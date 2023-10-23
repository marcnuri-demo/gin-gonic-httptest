package router

import (
	"encoding/json"
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
	if c.Request.ContentLength == 0 {
		c.IndentedJSON(400, "Empty body")
		return
	}
	if !containsHeader(c)("Content-Type", "application/json") {
		c.IndentedJSON(400, "Invalid Content-Type")
		return
	}
	data := make(map[string]interface{})
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		c.IndentedJSON(400, "Invalid JSON body")
		return
	}
	id, _ := uuid.NewRandom()
	data["id"] = id.String()
	entries.Set(id.String(), data)
	c.IndentedJSON(201, data)
}
