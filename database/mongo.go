package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-meli/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	// crea un cliente de mongodb con la configuración del .env
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// se ejecuta al final para cancelar el contexto y liberar recursos
	defer cancel()
	// crea las opciones de conexión con la URI del .env y ajusta el pool de conexiones
	opts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(50).
		SetMinPoolSize(5)
	// conecta al cliente de mongodb usando las opciones configuradas
	client, err := mongo.Connect(ctx, opts)
	// si hay un error al conectar, devuelve nil y el error
	if err != nil {
		return nil, fmt.Errorf("conectando a mongodb: %w", err)
	}
	// si el ping a la base de datos falla, devuelve nil y el error
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping a mongodb fallido: %w", err)
	}

	log.Println("✅ conectado a mongodb")
	// devuelve una instancia de MongoDB con el cliente y la base de datos conectados
	return &MongoDB{
		Client:   client,
		Database: client.Database(cfg.DB), // nombre de tu base de datos
	}, nil
}

// Disconnect cierra la conexión limpiamente — se llama con defer en main.go
func (m *MongoDB) Disconnect(ctx context.Context) {
	if err := m.Client.Disconnect(ctx); err != nil {
		log.Printf("error desconectando mongodb: %v", err)
		return
	}
	log.Println("🔌 mongodb desconectado")
}

// Collection devuelve una colección por nombre — lo usan los repositorios
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
