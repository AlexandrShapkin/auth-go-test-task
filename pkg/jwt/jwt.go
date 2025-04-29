package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)
// TODO: задокументировать методы + тесты
// TODO: вынести в отдельный файл и мб добавить обертки для не моих ошибок
var (
	InvalidTokenErr      = errors.New("invalid token")
	UnknownClaimsTypeErr = errors.New("unknown claims type")
	TokensNotPairedErr   = errors.New("tokens is not paired")
)

var (
	AccessExpires     = 30 * time.Minute
	RefreshExpires    = 7 * 24 * time.Hour
	ParseLeewayWindow = 10 * time.Second
)

type ImplJWT struct {
	accessSecretKey   []byte
	refreshSecretKey  []byte
	accessExpires     time.Duration
	refreshExpires    time.Duration
	parseLeewayWindow time.Duration
}

type JWT interface {
	GenereteTokenPair(uid string, userIP string) (string, string, error)
	ValidateAccessToken(accessToken string) (*AccessClaims, error)
	ValidateRefreshToken(refreshToken string) (*RefreshClaims, error)
	RefreshTokenPair(accessClaims *AccessClaims, refreshClaims *RefreshClaims) (string, string, error)
}

func NewJWT(
	accessSecretKey []byte,
	refreshSecretKey []byte,
	accessExpires time.Duration,
	refreshExpires time.Duration,
	parseLeewayWindow time.Duration,
) JWT {
	return &ImplJWT{
		accessSecretKey:   accessSecretKey,
		refreshSecretKey:  refreshSecretKey,
		accessExpires:     accessExpires,
		refreshExpires:    refreshExpires,
		parseLeewayWindow: parseLeewayWindow,
	}
}

type AccessClaims struct {
	UserIP string `json:"user_ip"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	AccessID string `json:"access_id"`
	UserIP   string `json:"user_ip"`
	jwt.RegisteredClaims
}

func (j *ImplJWT) GenereteTokenPair(uid string, userIP string) (string, string, error) {
	tokenID := uuid.NewString()

	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessClaims{
		UserIP: userIP,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uid,
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, RefreshClaims{
		AccessID: tokenID,
		UserIP:   userIP,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uid,
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	access, err := accessJWT.SignedString(j.accessSecretKey)
	if err != nil {
		return "", "", err
	}

	refresh, err := refreshJWT.SignedString(j.refreshSecretKey)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (j *ImplJWT) ValidateAccessToken(accessToken string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.accessSecretKey, nil
	}, jwt.WithLeeway(j.parseLeewayWindow))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, InvalidTokenErr
	}

	if claims, ok := token.Claims.(*AccessClaims); ok {
		return claims, nil
	}

	return nil, UnknownClaimsTypeErr
}

func (j *ImplJWT) ValidateRefreshToken(refreshToken string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.refreshSecretKey, nil
	}, jwt.WithLeeway(j.parseLeewayWindow))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, InvalidTokenErr
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok {
		return claims, nil
	}

	return nil, UnknownClaimsTypeErr
}

func (j *ImplJWT) RefreshTokenPair(accessClaims *AccessClaims, refreshClaims *RefreshClaims) (string, string, error) {
	if accessClaims.ID != refreshClaims.AccessID {
		return "", "", TokensNotPairedErr
	}

	return j.GenereteTokenPair(refreshClaims.Subject, refreshClaims.UserIP)
}
