package router

import (
	"go-meli/config"
	"go-meli/internal/handler"
	"go-meli/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func Setup(
	productoHandler *handler.ProductoHandler,
	authHandler *handler.AuthHandler,
	cfg *config.Config,
	log *zap.Logger,
) *gin.Engine {
	r := gin.New()

	// middlewares globales — aplican a todas las rutas
	limiter := middleware.NewRateLimiter(rate.Limit(cfg.RateLimit), cfg.RateBurst)
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(limiter.Middleware())

	// sub-routers
	registerAuthRoutes(r, authHandler)
	registerProductoRoutes(r, productoHandler, cfg)

	return r
}
