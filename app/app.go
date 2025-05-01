package app

import (
	"net/http"

	"github.com/AlexandrShapkin/auth-go-test-task/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type ImplApp struct {
	JWTManager jwt.JWT
	Router     *gin.Engine
}

type App interface {
	Run(addr string) error
}

func NewApp(jwtManager jwt.JWT) App {
	app := &ImplApp{
		JWTManager: jwtManager,
		Router:     gin.Default(),
	}
	// TODO: разработать схему бд и добавить работу с репозиторием бд в обработчики
	// TODO: сделать нормальную обработку ошибок и нормальные коды возврата
	app.Router.POST("/login/:guid", app.LoginHandler)
	app.Router.POST("/refresh", app.RefreshHandler)

	return app
}

func (a *ImplApp) Run(addr string) error {
	return a.Router.Run(addr) // TODO: адрес из конфиг файла
}

func (a *ImplApp) LoginHandler(ctx *gin.Context) {
	// TODO: добавить настройку для двух режимов, в одном вызывается ctx.ClientIP() для тестирования, а в другом ctx.RemoteIP()
	accessToken, refreshToken, err := a.JWTManager.GenereteTokenPair(ctx.Param("guid"), ctx.ClientIP()) // добавить проверку существования пользователя с полученным guid в бд иначе 404
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	// TODO: запись refreshToken в бд
	ctx.JSON(http.StatusOK, gin.H{ // TODO: запись в куки
		"access":  accessToken,
		"refresh": refreshToken,
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
	refreshClims, err := a.JWTManager.ValidateRefreshToken(body.Refresh)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	// TODO: проверка существования refreshToken в бд
	if accessClaims.UserIP != ctx.ClientIP() && refreshClims.UserIP != ctx.ClientIP() {
		// Такой подход к проверке адреса является ненадежным, как как в данном случае
		// ctx.ClientIP() вернет не действительный IP, а поле заголовка запроса X-Forwarded-For
		// Для получения действительного адреса можно вызвать ctx.RemoteIP()
		// TODO: добавить настройку для двух режимов, в одном вызывается ctx.ClientIP() для тестирования, а в другом ctx.RemoteIP()
		ctx.String(http.StatusBadRequest, "ip error")
		return
	}
	accessToken, refreshToken, err := a.JWTManager.RefreshTokenPair(accessClaims, refreshClims)
	ctx.JSON(http.StatusOK, gin.H{ // TODO: запись в куки
		"access":  accessToken,
		"refresh": refreshToken,
	})
}
