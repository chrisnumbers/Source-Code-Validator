package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/validate", GetValidateSource)

	err := router.Run("localhost:8080")
	if err != nil {
		return nil
	}

	return router
}
