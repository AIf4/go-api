package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter agrupa el limiter con el último tiempo de acceso
// — el tiempo de acceso sirve para limpiar IPs inactivas
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter mantiene un limiter por IP
type RateLimiter struct {
	limiters map[string]*ipLimiter
	mu       sync.RWMutex
	rate     rate.Limit // tokens por segundo
	burst    int        // capacidad máxima del balde
}

// NewRateLimiter crea el middleware
// r = requests por segundo permitidos
// b = burst máximo (capacidad del balde)
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		rate:     r,
		burst:    b,
	}

	// goroutine que limpia IPs inactivas cada minuto
	go rl.cleanupLoop()

	return rl
}

// getLimiter devuelve el limiter para una IP, creándolo si no existe
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	// primero intenta con RLock (solo lectura — múltiples goroutines pueden leer)
	rl.mu.RLock()
	il, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		il.lastSeen = time.Now()
		return il.limiter
	}

	// no existe — necesita escribir, usa Lock exclusivo
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// double-check: otra goroutine pudo haberlo creado mientras esperábamos el Lock
	if il, exists = rl.limiters[ip]; exists {
		return il.limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[ip] = &ipLimiter{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// cleanupLoop elimina IPs que no han hecho requests en los últimos 3 minutos
// evita que el map crezca indefinidamente en memoria
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, il := range rl.limiters {
			if time.Since(il.lastSeen) > 3*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware es la función que Gin ejecuta en cada request
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "demasiadas requests, intenta más tarde",
			})
			c.Abort() // corta la cadena — el handler no se ejecuta
			return
		}

		c.Next() // continúa al siguiente middleware o handler
	}
}
