package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-meli/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"created_at"`
}

type mongoUserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(col *mongo.Collection) UserRepository {
	return &mongoUserRepository{collection: col}
}

func (r *mongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	doc := userDoc{
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: time.Now(),
	}

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("creando usuario: %w", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *mongoUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var doc userDoc
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscando usuario: %w", err)
	}

	return &domain.User{
		ID:       doc.ID.Hex(),
		Email:    doc.Email,
		Password: doc.Password,
	}, nil
}
