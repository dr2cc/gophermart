package service

import (
	"crypto/sha1"
	"errors"
	"fmt"

	"gophermart/internal/models"
	"gophermart/internal/repository"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	salt       = "hjqrhjqw124617ajfhajs"
	signingKey = "qrkjk#4#%35FSFJlja#4353KSFjH"
	tokenTTL   = 12 * time.Hour
)

// tokenClaims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

// 📍Service implementation struct. Она же:
// - 📍Provider - в контексте Dependency Injection.
// Структура, которая «предоставляет» определенный функционал.
// - 📍Concrete Type. Технический термин, противопоставляющий структуру интерфейсу.
// - 📍Receiver. Такую структуру называют получателем методов,
// т.к. методы службы «привязаны» к этой структуре.
// Отдадим, в конструкторе ниже, структуру в которую там же
// Приняли интерфейс репозитория (для "общения" с базой).
type authService struct {
	// зависимости:
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *authService {
	return &authService{repo: repo}
}

// Внедрим (в структуру AuthService) метод CreateUser..
// В нем мы будем передавать пользователя, еще на слой ниже- в репозиторий.
func (s *authService) CreateUser(user models.User) (int, error) {
	user.Password = generatePasswordHash(user.Password)
	return s.repo.CreateUser(user)
}

func (s *authService) GenerateToken(login, password string) (string, error) {
	user, err := s.repo.GetUser(login, generatePasswordHash(password))
	if err != nil {
		return "", err
	}
	// создаём новый токен с алгоритмом подписи HS256
	// и утверждениями — tokenClaims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})

	return token.SignedString([]byte(signingKey))
}

func (s *authService) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
