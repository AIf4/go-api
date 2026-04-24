package router

import (
	"go-meli/config"
	"go-meli/internal/handler"
	"go-meli/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerProductoRoutes(r *gin.Engine, productoHandler *handler.ProductoHandler, cfg *config.Config) {
	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(cfg)) // JWT solo para estas rutas
	{
		productos := v1.Group("/productos")
		{
			productos.GET("", productoHandler.GetAll)
			productos.GET("/:id", productoHandler.GetByID)
			productos.POST("", productoHandler.Create)
			productos.POST("/many", productoHandler.InsertMany)
			productos.PUT("/:id", productoHandler.Update)
			productos.DELETE("/:id", productoHandler.Delete)
		}
	}
}
