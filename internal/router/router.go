package router

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/orcaman/concurrent-map/v2"
)

var entries = cmap.New[map[string]interface{}]()

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/", fallbackGet)
	router.POST("/", post)
	return router
}

func addCommonHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store")
	c.Header("Server", "gin-gonic/1.33.7")
}

func fallbackGet(c *gin.Context) {
	addCommonHeaders(c)
	c.IndentedJSON(200, "Cocktail service")
}

func post(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.IndentedJSON(400, "Empty body")
		return
	}
	if c.GetHeader("Content-Type") != "application/json" {
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
	addCommonHeaders(c)
	c.IndentedJSON(201, data)
}
