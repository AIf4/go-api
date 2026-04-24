// internal/middleware/jwt.go
package middleware

import (
	"net/http"
	"strings"

	"go-meli/config"
	"go-meli/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const ClaimsKey = "claims"

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// extrae el token del header Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "formato inválido: Bearer <token>"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// parsea y valida el token
		claims := &service.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			// verifica que el algoritmo sea el esperado
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido o expirado"})
			c.Abort()
			return
		}

		// guarda los claims en el contexto para que los handlers los usen
		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// GetClaims es un helper para extraer los claims desde cualquier handler
func GetClaims(c *gin.Context) *service.Claims {
	claims, _ := c.Get(ClaimsKey)
	return claims.(*service.Claims)
}
