package domain

import "errors"

type User struct {
	ID       string
	Email    string
	Password string // hash bcrypt — nunca el password plano
}

var (
	ErrEmailRequerido    = errors.New("email es requerido")
	ErrPasswordRequerido = errors.New("password es requerido")
	ErrCredenciales      = errors.New("credenciales inválidas")
	ErrUsuarioExiste     = errors.New("el email ya está registrado")
)
