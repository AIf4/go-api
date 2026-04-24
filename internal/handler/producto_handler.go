package handler

import (
	"errors"
	"fmt"
	"net/http"

	"go-meli/internal/domain"
	"go-meli/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductoHandler struct {
	service service.ProductoService
}

func NewProductoHandler(svc service.ProductoService) *ProductoHandler {
	return &ProductoHandler{service: svc}
}

// --- Request / Response structs — los tags json viven SOLO aquí ---

type createProductoRequest struct {
	Name            string            `json:"name"             binding:"required"`
	ImageURL        string            `json:"image_url"`
	Description     string            `json:"description"`
	Price           float64           `json:"price"            binding:"required,gt=0"`
	Size            string            `json:"size"`
	Weight          string            `json:"weight"`
	Color           string            `json:"color"`
	Brand           *string           `json:"brand,omitempty"`
	ModelVersion    *string           `json:"model_version,omitempty"`
	OS              *string           `json:"os,omitempty"`
	BatteryCapacity *string           `json:"battery_capacity,omitempty"`
	Camera          *string           `json:"camera,omitempty"`
	Memory          *string           `json:"memory,omitempty"`
	Storage         *string           `json:"storage,omitempty"`
	Specifications  map[string]string `json:"specifications,omitempty"`
}

type updateProductoRequest struct {
	Name            string            `json:"name"             binding:"required"`
	ImageURL        string            `json:"image_url"`
	Description     string            `json:"description"`
	Price           float64           `json:"price"            binding:"required,gt=0"`
	Size            string            `json:"size"`
	Weight          string            `json:"weight"`
	Color           string            `json:"color"`
	Brand           *string           `json:"brand,omitempty"`
	ModelVersion    *string           `json:"model_version,omitempty"`
	OS              *string           `json:"os,omitempty"`
	BatteryCapacity *string           `json:"battery_capacity,omitempty"`
	Camera          *string           `json:"camera,omitempty"`
	Memory          *string           `json:"memory,omitempty"`
	Storage         *string           `json:"storage,omitempty"`
	Specifications  map[string]string `json:"specifications,omitempty"`
}

type productoResponse struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	ImageURL        string            `json:"image_url"`
	Description     string            `json:"description"`
	Price           float64           `json:"price"`
	Size            string            `json:"size"`
	Weight          string            `json:"weight"`
	Color           string            `json:"color"`
	Rating          float64           `json:"rating"`
	Brand           *string           `json:"brand,omitempty"`
	ModelVersion    *string           `json:"model_version,omitempty"`
	OS              *string           `json:"os,omitempty"`
	BatteryCapacity *string           `json:"battery_capacity,omitempty"`
	Camera          *string           `json:"camera,omitempty"`
	Memory          *string           `json:"memory,omitempty"`
	Storage         *string           `json:"storage,omitempty"`
	Specifications  map[string]string `json:"specifications,omitempty"`
}

type insertManyResponse struct {
	Inserted int                `json:"inserted"`
	Errors   []insertManyError  `json:"errors,omitempty"`
	Items    []productoResponse `json:"items"`
}

type insertManyError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

// --- Handlers ---

func (h *ProductoHandler) GetAll(c *gin.Context) {
	productos, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error obteniendo productos"})
		return
	}

	response := make([]productoResponse, len(productos))
	for i, p := range productos {
		response[i] = toResponse(p)
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductoHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	producto, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{"error": "producto no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error obteniendo producto"})
		return
	}

	c.JSON(http.StatusOK, toResponse(producto))
}

func (h *ProductoHandler) Create(c *gin.Context) {
	var req createProductoRequest

	// binding:"required" y binding:"gt=0" son validados aquí automáticamente
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.CreateProductoCmd{
		Name:            req.Name,
		ImageURL:        req.ImageURL,
		Description:     req.Description,
		Price:           req.Price,
		Size:            req.Size,
		Weight:          req.Weight,
		Color:           req.Color,
		Brand:           req.Brand,
		ModelVersion:    req.ModelVersion,
		OS:              req.OS,
		BatteryCapacity: req.BatteryCapacity,
		Camera:          req.Camera,
		Memory:          req.Memory,
		Storage:         req.Storage,
		Specifications:  req.Specifications,
	}

	producto, err := h.service.Create(c.Request.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrNameRequerido) || errors.Is(err, domain.ErrPrecioInvalido) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creando producto"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(producto))
}

func (h *ProductoHandler) InsertMany(c *gin.Context) {
	// recibe un array de createProductoRequest directamente
	var req []createProductoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lista vacía"})
		return
	}

	// convierte []createProductoRequest → []CreateProductoCmd
	cmds := make([]service.CreateProductoCmd, len(req))
	for i, p := range req {
		cmds[i] = service.CreateProductoCmd{
			Name:            p.Name,
			ImageURL:        p.ImageURL,
			Description:     p.Description,
			Price:           p.Price,
			Size:            p.Size,
			Weight:          p.Weight,
			Color:           p.Color,
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

	// llama al service
	productos, prodErrors := h.service.InsertMany(c.Request.Context(), cmds)
	if prodErrors != nil {
		errMap := make(map[string]string, len(prodErrors))
		for i, err := range prodErrors {
			errMap[fmt.Sprintf("index_%d", i)] = err.Error()
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":  "algunos productos fallaron",
			"detail": errMap,
		})
		return
	}

	// convierte a response
	response := make([]productoResponse, len(productos))
	for i, p := range productos {
		response[i] = toResponse(p)
	}

	c.JSON(http.StatusCreated, gin.H{
		"inserted": len(response),
		"products": response,
	})
}

func (h *ProductoHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req updateProductoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.UpdateProductoCmd{
		ID:              id,
		Name:            req.Name,
		ImageURL:        req.ImageURL,
		Description:     req.Description,
		Price:           req.Price,
		Size:            req.Size,
		Weight:          req.Weight,
		Color:           req.Color,
		Brand:           req.Brand,
		ModelVersion:    req.ModelVersion,
		OS:              req.OS,
		BatteryCapacity: req.BatteryCapacity,
		Camera:          req.Camera,
		Memory:          req.Memory,
		Storage:         req.Storage,
		Specifications:  req.Specifications,
	}

	producto, err := h.service.Update(c.Request.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrNoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{"error": "producto no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error actualizando producto"})
		return
	}

	c.JSON(http.StatusOK, toResponse(producto))
}

func (h *ProductoHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{"error": "producto no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error eliminando producto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "producto eliminado"})
}

// --- mapper: domain.Producto → productoResponse ---

func toResponse(p *domain.Producto) productoResponse {
	return productoResponse{
		ID:              p.ID,
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

// ---
