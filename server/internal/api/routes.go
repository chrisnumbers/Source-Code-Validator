package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"source-code-validator/server/internal/util"
)

func SetupRouter(handler *util.Handler) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	router.POST("/validate", func(c *gin.Context) {
		PostValidateSource(c, handler)
	})

	err := router.Run("localhost:8080")
	if err != nil {
		return nil
	}

	return router
}
