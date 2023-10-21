package main

import (
	"fmt"
	"github.com/marcnuri-demo/gin-gonic-httptest/internal/router"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error: ", r)
			os.Exit(1)
		}
	}()
	err := router.SetupRouter().Run("0.0.0.0:8080")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
