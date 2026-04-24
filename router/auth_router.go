package router

import (
	"go-meli/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}
