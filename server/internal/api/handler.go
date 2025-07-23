package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"source-code-validator/server/internal/service"
	"source-code-validator/server/internal/util"
)

func PostValidateSource(c *gin.Context, handler *util.Handler) {
	url := c.PostForm("url")
	fmt.Println(url)

	requirements, err := c.FormFile("requirements")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "failed to validate"})
		return
	}

	consultation, err := service.ValidateSourceCode(url, requirements, handler)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("failed to validate: %v", err)})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": consultation})

}
