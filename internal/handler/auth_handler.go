// internal/handler/auth_handler.go
package handler

import (
	"errors"
	"net/http"

	"go-meli/internal/domain"
	"go-meli/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

type registerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register godoc
// @Summary     Registrar usuario
// @Description Crea un nuevo usuario con email y password
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body     registerRequest true "Datos de registro"
// @Success     201  {object} map[string]string "usuario registrado"
// @Failure     400  {object} map[string]string "error de validación"
// @Failure     409  {object} map[string]string "email ya registrado"
// @Failure     500  {object} map[string]string "error interno"
// @Router      /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Register(c.Request.Context(), service.RegisterCmd{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUsuarioExiste) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error registrando usuario"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "usuario registrado"})
}

// Login godoc
// @Summary     Iniciar sesión
// @Description Autentica un usuario y devuelve un token JWT
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body     loginRequest true "Credenciales"
// @Success     200  {object} map[string]string "token JWT"
// @Failure     400  {object} map[string]string "error de validación"
// @Failure     401  {object} map[string]string "credenciales inválidas"
// @Failure     500  {object} map[string]string "error interno"
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(c.Request.Context(), service.LoginCmd{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCredenciales) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales inválidas"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error iniciando sesión"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
