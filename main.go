// @title           Go-Meli API
// @version         1.0
// @description     API REST para gestión de productos estilo MercadoLibre.
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
// @description Ingresa el token JWT con el prefijo "Bearer ". Ejemplo: "Bearer eyJhbGci..."

package main

import (
	"context"
	stdlog "log" // alias para evitar conflicto con la variable zap
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-meli/config"
	"go-meli/database"
	"go-meli/internal/handler"
	"go-meli/internal/logger"
	"go-meli/internal/repository"
	"go-meli/internal/service"
	"go-meli/router"

	_ "go-meli/docs"

	"go.uber.org/zap"
)

func main() {
	// 1. config
	cfg := config.LoadConfig()

	// 2. logger — primero que todo para poder loguear errores de arranque
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		stdlog.Fatalf("error creando directorio de logs: %v", err)
	}

	log, err := logger.New(cfg)
	if err != nil {
		stdlog.Fatalf("error iniciando logger: %v", err)
	}
	defer log.Sync()

	// 3. base de datos
	db, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatal("error iniciando mongodb", zap.Error(err))
	}

	// 4. capas
	productoRepo := repository.NewProductoRepository(db.Collection("productos"))
	productoService := service.NewProductoService(productoRepo)
	productoHandler := handler.NewProductoHandler(productoService)
	//
	userRepo := repository.NewUserRepository(db.Collection("users"))
	authService := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// 5. router
	r := router.Setup(productoHandler, authHandler, cfg, log)

	// 6. servidor HTTP
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// 7. graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("servidor corriendo", zap.String("puerto", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("error iniciando servidor", zap.Error(err))
		}
	}()

	// 8. espera señal de cierre
	<-ctx.Done()
	log.Info("apagando servidor...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("error en shutdown", zap.Error(err))
	}

	// 9. mongo se cierra después del servidor
	db.Disconnect(shutdownCtx)

	log.Info("servidor apagado correctamente")
}
