// internal/service/auth_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"go-meli/config"
	"go-meli/internal/domain"
	"go-meli/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, cmd RegisterCmd) error
	Login(ctx context.Context, cmd LoginCmd) (string, error)
}

type RegisterCmd struct {
	Email    string
	Password string
}

type LoginCmd struct {
	Email    string
	Password string
}

// Claims son los datos que viven dentro del token
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(repo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{userRepo: repo, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, cmd RegisterCmd) error {
	if cmd.Email == "" {
		return domain.ErrEmailRequerido
	}
	if cmd.Password == "" {
		return domain.ErrPasswordRequerido
	}

	// verifica que el email no esté registrado
	existing, err := s.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return domain.ErrUsuarioExiste
	}

	// hashea el password — nunca guardes el password plano
	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hasheando password: %w", err)
	}

	user := &domain.User{
		Email:    cmd.Email,
		Password: string(hash),
	}

	return s.userRepo.Create(ctx, user)
}

func (s *authService) Login(ctx context.Context, cmd LoginCmd) (string, error) {
	if cmd.Email == "" || cmd.Password == "" {
		return "", domain.ErrCredenciales
	}

	// busca el usuario
	user, err := s.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return "", err
	}

	// nil significa que no existe — mismo error que password incorrecto
	// nunca digas "el email no existe" — das información al atacante
	if user == nil {
		return "", domain.ErrCredenciales
	}

	// compara el password con el hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.Password)); err != nil {
		return "", domain.ErrCredenciales
	}

	// genera el token
	token, err := s.generateToken(user)
	if err != nil {
		return "", fmt.Errorf("generando token: %w", err)
	}

	return token, nil
}

func (s *authService) generateToken(user *domain.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
