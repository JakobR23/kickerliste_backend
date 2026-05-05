package router

import (
	"net/http"

	"bierliste_backend/internal/entity"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	router := gin.Default()
	router.GET("/", test)
	return router
}

func test(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, entity.User{})
}
