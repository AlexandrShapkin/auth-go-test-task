package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// Значение которое будет прибавлено к текущему времени, чтобы установить время когда access токен истечет
	AccessExpires = 30 * time.Minute
	// Значение которое будет прибавлено к текущему времени, чтобы установить время когда refresh токен истечет
	RefreshExpires = 7 * 24 * time.Hour
	// Значние допустимого расхождения времени при проверке действительности токена
	ParseLeewayWindow = 10 * time.Second
)

type ImplJWT struct {
	accessSecretKey   []byte
	refreshSecretKey  []byte
	accessExpires     time.Duration
	refreshExpires    time.Duration
	parseLeewayWindow time.Duration
}

// Менеджер работы с токенами
type JWT interface {
	// Создает новую пару токенов исходя из uuid пользователя и IP адреса с которого был сделан запрос на получение
	//
	// Оба токена имеют тип JWT. access (первый) - HS512, refresh (второй) - HS256
	//
	// uuid пользовтаеля записывается в поле Subject (sub), userIP в UserIP (user_ip)
	GenereteTokenPair(uid string, userIP string) (string, string, error)
	// Проверяет действительность access токена, в случае если токен действителен, возвращает его payload
	ValidateAccessToken(accessToken string) (*AccessClaims, error)
	// Проверяет действительность refresh токена, в случае если токен действителен, возвращает его payload
	ValidateRefreshToken(refreshToken string) (*RefreshClaims, error)
	// Обертка вокруг GenereteTokenPair, но проверяет связанность токенов
	RefreshTokenPair(accessClaims *AccessClaims, refreshClaims *RefreshClaims, currentUserIP string) (string, string, error)
	// Возвращает время в течении которого access токен валиден с момента создания
	GetAccessExpires() time.Duration
	// Возвращает время в течении которого access токен валиден с момента создания в секундах
	GetAccessExpiresSec() int
	// Возвращает время в течении которого refresh токен валиден с момента создания
	GetRefreshExpires() time.Duration
	// Возвращает время в течении которого refresh токен валиден с момента создания в секундах
	GetRefreshExpiresSec() int
}

// Конструктор менеджера токенов. Более предпочтительно чем создавать из голой структуры
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

// Payload access токена
type AccessClaims struct {
	// IP с которого был выполнен запрос на получение токена
	UserIP string `json:"user_ip"`
	jwt.RegisteredClaims
}

// Payload refresh токена
type RefreshClaims struct {
	// ID (jti) связанного access токена
	AccessID string `json:"access_id"`
	// IP с которого был выполнен запрос на получение токена
	UserIP string `json:"user_ip"`
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
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*AccessClaims); ok {
		return claims, nil
	}

	return nil, ErrUnknownClaimsType
}

func (j *ImplJWT) ValidateRefreshToken(refreshToken string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.refreshSecretKey, nil
	}, jwt.WithLeeway(j.parseLeewayWindow))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok {
		return claims, nil
	}

	return nil, ErrUnknownClaimsType
}

func (j *ImplJWT) RefreshTokenPair(accessClaims *AccessClaims, refreshClaims *RefreshClaims, currentUserIP string) (string, string, error) {
	if accessClaims.ID != refreshClaims.AccessID {
		return "", "", ErrTokensNotPaired
	}

	return j.GenereteTokenPair(refreshClaims.Subject, currentUserIP)
}

func (j *ImplJWT) GetAccessExpires() time.Duration {
	return j.accessExpires
}

func (j *ImplJWT) GetAccessExpiresSec() int {
	return int(j.accessExpires.Seconds())
}

func (j *ImplJWT) GetRefreshExpires() time.Duration {
	return j.refreshExpires
}

func (j *ImplJWT) GetRefreshExpiresSec() int {
	return int(j.refreshExpires.Seconds())
}
