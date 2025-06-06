package services

import (
	"context"
	"errors"
	"time"

	"billing-system/internal/models"
	"billing-system/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
	ValidateToken(tokenString string) (*models.Claims, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

type authService struct {
	userRepo  repositories.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: []byte("billing-secret-key-change-in-production"), // В продакшене использовать переменную окружения
	}
}

func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Проверяем существует ли пользователь
	existingUser, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, errors.New("пользователь с таким именем уже существует")
	}

	existingEmail, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingEmail != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("ошибка при хешировании пароля")
	}

	// Создаем пользователя
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("ошибка при создании пользователя")
	}

	// Генерируем токен
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("ошибка при генерации токена")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// Ищем пользователя
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("неверные учетные данные")
	}

	// Проверяем активен ли пользователь
	if !user.IsActive {
		return nil, errors.New("аккаунт заблокирован")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("неверные учетные данные")
	}

	// Обновляем время последнего входа
	s.userRepo.UpdateLastLogin(ctx, user.ID.Hex())

	// Генерируем токен
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("ошибка при генерации токена")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) ValidateToken(tokenString string) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("недействительный токен")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, errors.New("недействительный токен")
	}

	userID, ok := (*claims)["user_id"].(string)
	if !ok {
		return nil, errors.New("недействительный токен")
	}

	username, ok := (*claims)["username"].(string)
	if !ok {
		return nil, errors.New("недействительный токен")
	}

	role, ok := (*claims)["role"].(string)
	if !ok {
		return nil, errors.New("недействительный токен")
	}

	return &models.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, nil
}

func (s *authService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *authService) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
