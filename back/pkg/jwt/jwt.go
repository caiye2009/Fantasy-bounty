package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims JWT Claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTService JWT服务
type JWTService struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

// NewJWTService 创建JWT服务
func NewJWTService(secretKey, issuer string, expiry time.Duration) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		expiry:    expiry,
	}
}

// GenerateToken 生成JWT token
func (s *JWTService) GenerateToken(username string) (string, error) {
	now := time.Now()
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken 验证JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ParseTokenIgnoreExpiry 解析 token，忽略过期错误但验证签名
// 用于 refresh 场景：token 过期但签名有效时，允许换发新 token
func (s *JWTService) ParseTokenIgnoreExpiry(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
