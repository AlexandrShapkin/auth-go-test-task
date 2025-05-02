package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/AlexandrShapkin/auth-go-test-task/pkg/jwt"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/mailer"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/models"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/repositories"
	"github.com/gin-gonic/gin"
)

const (
	RefreshTokenName = "refresh_token"
	AccessTokenName  = "access_token"
)

type ImplApp struct {
	JWTManager          jwt.JWT
	UserRepo            repositories.UserRepo
	Mailer              mailer.Mailer
	Router              *gin.Engine
	LoginRemoteIPMode   bool
	RefreshRemoteIPMode bool
	Domain              string
}

type App interface {
	Run(addr string) error
}

func NewApp(
	jwtManager jwt.JWT,
	userRepo repositories.UserRepo,
	mailer mailer.Mailer,
	loginRemoteIPMode bool,
	refreshRemoteIPMode bool,
	domain string,
) App {
	app := &ImplApp{
		JWTManager:          jwtManager,
		UserRepo:            userRepo,
		Mailer:              mailer,
		Router:              gin.Default(),
		LoginRemoteIPMode:   loginRemoteIPMode,
		RefreshRemoteIPMode: refreshRemoteIPMode,
		Domain:              domain,
	}
	// TODO: сделать нормальную обработку ошибок и нормальные коды возврата
	app.Router.POST("/register", app.RegisterHandler)
	app.Router.POST("/login/:guid", app.LoginHandler)
	app.Router.POST("/refresh", app.RefreshHandler)

	return app
}

func (a *ImplApp) Run(addr string) error {
	return a.Router.Run(addr)
}

type LoginBody struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Обработчик регистрации пользователя, сделан в упрощенном виде, т.к. нужен в основном для проверки работы
func (a *ImplApp) RegisterHandler(ctx *gin.Context) {
	body := LoginBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidRequestData.Error()})
		return
	}

	hashedPassword, err := HashPassword(body.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user := models.NewUser(body.Email, hashedPassword)

	err = a.UserRepo.Create(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": MessageSuccessfullyRegistered,
		"user_id": user.UserID.String(),
	})
}

func (a *ImplApp) LoginHandler(ctx *gin.Context) {
	body := LoginBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidRequestData.Error()})
		return
	}

	user, err := a.UserRepo.FindByIDString(ctx, ctx.Param("guid"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": ErrUserNotFound.Error()})
		return
	}

	if user.Email != body.Email || !CompareHashAndPassword(user.Password, body.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidCredentials.Error()})
		return
	}

	// ctx.ClientIP() вернет не действительный IP, а поле заголовка запроса X-Forwarded-For
	// Для получения действительного адреса можно вызвать ctx.RemoteIP()
	clientIP := ctx.ClientIP()
	if a.LoginRemoteIPMode {
		clientIP = ctx.RemoteIP()
	}
	accessToken, refreshToken, err := a.JWTManager.GenereteTokenPair(ctx.Param("guid"), clientIP) // добавить проверку существования пользователя с полученным guid в бд иначе 404
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	b64token := EncodeTokenToBase64(refreshToken)
	err = a.SaveRefreshToDB(ctx, b64token, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accessExpiresSec := a.JWTManager.GetAccessExpiresSec()
	refreshExpiresSec := a.JWTManager.GetRefreshExpiresSec()

	// здесь нет ошибки связанной с времени жизни access токена, так как он нужен для /refresh операции
	ctx.SetCookie(AccessTokenName, accessToken, refreshExpiresSec, "/", a.Domain, false, true)
	ctx.SetCookie(RefreshTokenName, b64token, refreshExpiresSec, "/", a.Domain, false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"message":    MessageSuccessfullyLoggedIn,
		"expires_in": accessExpiresSec, // в качестве времени истечения access токена хоть как отправляю время его валидности
	})
}

func (a *ImplApp) RefreshHandler(ctx *gin.Context) {
	accessToken, err := ctx.Cookie(AccessTokenName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": ErrAccessTokenRequired.Error()})
		return
	}

	refreshToken, err := ctx.Cookie(RefreshTokenName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": ErrRefreshTokenRequired.Error()})
		return
	}

	accessClaims, err := a.JWTManager.ValidateAccessToken(accessToken)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	rawToken, err := DecodeTokenFromBase64(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	refreshClims, err := a.JWTManager.ValidateRefreshToken(rawToken)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	user, err := a.UserRepo.FindByIDString(ctx, refreshClims.Subject)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": ErrUserNotFound})
		return
	}

	if !CompareHashAndToken(user.RefreshToken, refreshToken) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": ErrIncorrectRefreshToken.Error()})
		return
	}

	// ctx.ClientIP() вернет не действительный IP, а поле заголовка запроса X-Forwarded-For
	// Для получения действительного адреса можно вызвать ctx.RemoteIP()
	clientIP := ctx.ClientIP()
	if a.RefreshRemoteIPMode {
		clientIP = ctx.RemoteIP()
	}
	if accessClaims.UserIP != clientIP && refreshClims.UserIP != clientIP {
		go func() {
			err := a.Mailer.SendMail(
				user.Email,
				"Предупреждение о доступе к аккаунту с нового IP адреса",
				fmt.Sprintf("Доступ к аккаунту был выполнен с неавторизованного IP адреса (%s)\n"+
					"Если это не вы, то обратитесь к системному администратору", clientIP),
			)
			if err != nil {
				slog.Warn("Failed to send mail", "error", err.Error())
			}
		}()
	}

	accessToken, refreshToken, err = a.JWTManager.RefreshTokenPair(accessClaims, refreshClims, clientIP)

	b64token := EncodeTokenToBase64(refreshToken)
	err = a.SaveRefreshToDB(ctx, b64token, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accessExpiresSec := a.JWTManager.GetAccessExpiresSec()
	refreshExpiresSec := a.JWTManager.GetRefreshExpiresSec()

	ctx.SetCookie(AccessTokenName, accessToken, refreshExpiresSec, "/", a.Domain, false, true)
	ctx.SetCookie(RefreshTokenName, b64token, refreshExpiresSec, "/", a.Domain, false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"message":    MessafeSuccessfullyRefreshed,
		"expires_in": accessExpiresSec,
	})
}

func (a *ImplApp) SaveRefreshToDB(ctx context.Context, b64refresh string, user *models.User) error {
	hashedRefresh, err := HashToken(b64refresh)
	if err != nil {
		return err
	}
	user.RefreshToken = hashedRefresh

	err = a.UserRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	return nil
}
