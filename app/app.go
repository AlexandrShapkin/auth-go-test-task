package app

import (
	"context"
	"net/http"

	"github.com/AlexandrShapkin/auth-go-test-task/pkg/jwt"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/models"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/repositories"
	"github.com/gin-gonic/gin"
)

type ImplApp struct {
	JWTManager          jwt.JWT
	UserRepo            repositories.UserRepo
	Router              *gin.Engine
	LoginRemoteIPMode   bool
	RefreshRemoteIPMode bool
}

type App interface {
	Run(addr string) error
}

func NewApp(
	jwtManager jwt.JWT,
	userRepo repositories.UserRepo,
	loginRemoteIPMode bool,
	refreshRemoteIPMode bool,
) App {
	app := &ImplApp{
		JWTManager:          jwtManager,
		UserRepo:            userRepo,
		Router:              gin.Default(),
		LoginRemoteIPMode:   loginRemoteIPMode,
		RefreshRemoteIPMode: refreshRemoteIPMode,
	}
	// TODO: сделать нормальную обработку ошибок и нормальные коды возврата
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

func (a *ImplApp) LoginHandler(ctx *gin.Context) {
	body := LoginBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"}) // TODO: вынести в отдельный тип ошибки
		return
	}

	user, err := a.UserRepo.FindByIDString(ctx, ctx.Param("guid"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"}) // TODO: вынести в отдельный тип ошибки
		return
	}

	if user.Email != body.Email || user.Password != body.Password { // TODO: проверка хеша пароля
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect email or password"}) // TODO: вынести в отдельный тип ошибки
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // TODO: вынести в отдельный тип ошибки
		return
	}

	ctx.JSON(http.StatusOK, gin.H{ // TODO: запись в куки
		"access":  accessToken,
		"refresh": b64token,
	})
}

func (a *ImplApp) RefreshHandler(ctx *gin.Context) {
	type Body struct { // TODO: чтение из куки
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}
	body := Body{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	accessClaims, err := a.JWTManager.ValidateAccessToken(body.Access)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	rawToken, err := DecodeTokenFromBase64(body.Refresh)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // TODO: вынести в отдельный тип ошибки
		return
	}
	refreshClims, err := a.JWTManager.ValidateRefreshToken(rawToken)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	user, err := a.UserRepo.FindByIDString(ctx, refreshClims.Subject)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"}) // TODO: вынести в отдельный тип ошибки
		return
	}

	if !CompareHashAndToken(user.RefreshToken, body.Refresh) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect refresh token"})
		return
	}

	// ctx.ClientIP() вернет не действительный IP, а поле заголовка запроса X-Forwarded-For
	// Для получения действительного адреса можно вызвать ctx.RemoteIP()
	clientIP := ctx.ClientIP()
	if a.RefreshRemoteIPMode {
		clientIP = ctx.RemoteIP()
	}
	if accessClaims.UserIP != clientIP && refreshClims.UserIP != clientIP {
		ctx.String(http.StatusBadRequest, "ip error")
		return
	}

	accessToken, refreshToken, err := a.JWTManager.RefreshTokenPair(accessClaims, refreshClims)

	b64token := EncodeTokenToBase64(refreshToken)
	err = a.SaveRefreshToDB(ctx, b64token, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // TODO: вынести в отдельный тип ошибки
		return
	}

	ctx.JSON(http.StatusOK, gin.H{ // TODO: запись в куки
		"access":  accessToken,
		"refresh": b64token,
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
