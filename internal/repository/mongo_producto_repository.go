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

// productoDoc es el modelo de Mongo — los tags bson viven SOLO aquí
type productoDoc struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	Name            string             `bson:"name"`
	ImageURL        string             `bson:"image_url"`
	Description     string             `bson:"description"`
	Price           float64            `bson:"price"`
	Size            string             `bson:"size"`
	Weight          string             `bson:"weight"`
	Color           string             `bson:"color"`
	Rating          float64            `bson:"rating"`
	Brand           *string            `bson:"brand,omitempty"`
	ModelVersion    *string            `bson:"model_version,omitempty"`
	OS              *string            `bson:"os,omitempty"`
	BatteryCapacity *string            `bson:"battery_capacity,omitempty"`
	Camera          *string            `bson:"camera,omitempty"`
	Memory          *string            `bson:"memory,omitempty"`
	Storage         *string            `bson:"storage,omitempty"`
	Specifications  map[string]string  `bson:"specifications,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"`
}

type mongoProductoRepository struct {
	collection *mongo.Collection
}

func NewProductoRepository(col *mongo.Collection) ProductoRepository {
	return &mongoProductoRepository{collection: col}
}

func (r *mongoProductoRepository) FindAll(ctx context.Context) ([]*domain.Producto, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("buscando productos: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []productoDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decodificando productos: %w", err)
	}

	productos := make([]*domain.Producto, len(docs))
	for i, doc := range docs {
		productos[i] = toDomain(doc)
	}
	return productos, nil
}

func (r *mongoProductoRepository) FindByID(ctx context.Context, id string) (*domain.Producto, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrIDInvalido
	}

	var doc productoDoc
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscando producto: %w", err)
	}

	return toDomain(doc), nil
}

func (r *mongoProductoRepository) FindInIDs(ctx context.Context, ids []string) (*domain.FindInResult, error) {
	// convierte strings a ObjectIDs — los inválidos van directo a NotFound
	oids := make([]primitive.ObjectID, 0, len(ids))
	result := &domain.FindInResult{
		Found:    make([]*domain.Producto, 0),
		NotFound: make([]string, 0),
	}

	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			// ID con formato inválido — no puede existir en Mongo
			result.NotFound = append(result.NotFound, id)
			continue
		}
		oids = append(oids, oid)
	}

	// si todos los IDs eran inválidos, no hay nada que buscar
	if len(oids) == 0 {
		result.NotFound = ids
		return result, nil
	}

	// una sola query con todos los IDs válidos
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		return nil, fmt.Errorf("buscando productos: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []productoDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decodificando productos: %w", err)
	}

	// construye un set de IDs encontrados para comparar rápido
	foundIDs := make(map[string]bool, len(docs))
	for _, doc := range docs {
		producto := toDomain(doc)
		result.Found = append(result.Found, producto)
		foundIDs[producto.ID] = true
	}

	// compara los IDs solicitados contra los encontrados
	for _, id := range ids {
		if !foundIDs[id] {
			result.NotFound = append(result.NotFound, id)
		}
	}

	return result, nil
}

func (r *mongoProductoRepository) Create(ctx context.Context, producto *domain.Producto) error {
	doc := toDoc(producto)
	doc.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("creando producto: %w", err)
	}

	producto.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (s *mongoProductoRepository) InsertMany(ctx context.Context, productos []*domain.Producto) ([]*domain.Producto, map[int]error) {
	results := make([]*domain.Producto, 0, len(productos))
	// defino el slice
	docs := make([]interface{}, len(productos))
	// recorro los productos
	for i, p := range productos {
		// convierte domain → documento mongo
		doc := toDoc(p)
		doc.CreatedAt = time.Now()
		// tengo todos los documentos en el slice ya mapeados
		docs[i] = doc
	}

	result, err := s.collection.InsertMany(ctx, docs)
	// si hay error, devuelvo el error inmediatamente
	if err != nil {
		return nil, map[int]error{0: fmt.Errorf("insertando productos: %w", err)}
	}
	// recorro todos los resultados para asignar los IDs
	for i, id := range result.InsertedIDs {
		// se asigna el ID al producto correspondiente
		productos[i].ID = id.(primitive.ObjectID).Hex()
		// se asigna el producto a la lista de resultados
		results = append(results, productos[i])
	}
	// se devuelven los productos insertados con sus IDs asignados
	return results, nil
}

func (r *mongoProductoRepository) Update(ctx context.Context, producto *domain.Producto) error {
	oid, err := primitive.ObjectIDFromHex(producto.ID)
	if err != nil {
		return fmt.Errorf("id inválido: %w", err)
	}

	update := bson.M{"$set": bson.M{
		"name":             producto.Name,
		"image_url":        producto.ImageURL,
		"description":      producto.Description,
		"price":            producto.Price,
		"size":             producto.Size,
		"weight":           producto.Weight,
		"color":            producto.Color,
		"rating":           producto.Rating,
		"brand":            producto.Brand,
		"model_version":    producto.ModelVersion,
		"os":               producto.OS,
		"battery_capacity": producto.BatteryCapacity,
		"camera":           producto.Camera,
		"memory":           producto.Memory,
		"storage":          producto.Storage,
		"specifications":   producto.Specifications,
	}}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return fmt.Errorf("actualizando producto: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrNoEncontrado
	}

	return nil
}

func (r *mongoProductoRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("id inválido: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("eliminando producto: %w", err)
	}
	if result.DeletedCount == 0 {
		return domain.ErrNoEncontrado
	}

	return nil
}

// --- mappers ---

func toDomain(doc productoDoc) *domain.Producto {
	return &domain.Producto{
		ID:              doc.ID.Hex(),
		Name:            doc.Name,
		ImageURL:        doc.ImageURL,
		Description:     doc.Description,
		Price:           doc.Price,
		Size:            doc.Size,
		Weight:          doc.Weight,
		Color:           doc.Color,
		Rating:          doc.Rating,
		Brand:           doc.Brand,
		ModelVersion:    doc.ModelVersion,
		OS:              doc.OS,
		BatteryCapacity: doc.BatteryCapacity,
		Camera:          doc.Camera,
		Memory:          doc.Memory,
		Storage:         doc.Storage,
		Specifications:  doc.Specifications,
	}
}

func toDoc(p *domain.Producto) productoDoc {
	return productoDoc{
		Name:            p.Name,
		ImageURL:        p.ImageURL,
		Description:     p.Description,
		Price:           p.Price,
		Size:            p.Size,
		Weight:          p.Weight,
		Color:           p.Color,
		Rating:          p.Rating,
		Brand:           p.Brand,
		ModelVersion:    p.ModelVersion,
		OS:              p.OS,
		BatteryCapacity: p.BatteryCapacity,
		Camera:          p.Camera,
		Memory:          p.Memory,
		Storage:         p.Storage,
		Specifications:  p.Specifications,
	}
}
