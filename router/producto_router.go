package router

import (
	"go-meli/config"
	"go-meli/internal/handler"
	"go-meli/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerProductoRoutes(r *gin.Engine, productoHandler *handler.ProductoHandler, cfg *config.Config) {
	v1 := r.Group("/api/v1")

	// JWT solo para estas rutas
	//

	productosPublic := v1.Group("/productos")
	{
		productosPublic.GET("", productoHandler.GetAll)
		productosPublic.GET("/:id", productoHandler.GetByID)
		productosPublic.POST("/find-in", productoHandler.FindInIDs)
	}

	productosPrivate := v1.Group("/productos")
	productosPrivate.Use(middleware.JWTAuth(cfg))
	{
		productosPrivate.POST("", productoHandler.Create)
		productosPrivate.POST("/many", productoHandler.InsertMany)
		productosPrivate.PUT("/:id", productoHandler.Update)
		productosPrivate.DELETE("/:id", productoHandler.Delete)
	}

}
