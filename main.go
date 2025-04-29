package main

import (
	"net/http"

	"github.com/AlexandrShapkin/auth-go-test-task/pkg/jwt"
	"github.com/gin-gonic/gin"
)
// TODO: cobra? viper?  
func main() {
	jwtManager := jwt.NewJWT(
		// TODO: добавить загрузку параметров из конфиг файла (есть в каком-то проекте)
		[]byte("a-string-secret-at-least-256-bits-long"),
		[]byte("a-string-secret-at-least-256-bits-long"),
		jwt.AccessExpires,
		jwt.RefreshExpires,
		jwt.ParseLeewayWindow,
	)
	// TODO: вынести роутер в пакет app
	router := gin.Default()
	// TODO: разработать схему бд и добавить работу с репозиторием бд в обработчики
	// TODO: сделать нормальную обработку ошибок и нормальные коды возврата
	router.POST("/login/:guid", func(ctx *gin.Context) {
		// TODO: добавить настройку для двух режимов, в одном вызывается ctx.ClientIP() для тестирования, а в другом ctx.RemoteIP()
		accessToken, refreshToken, err := jwtManager.GenereteTokenPair(ctx.Param("guid"), ctx.ClientIP()) // добавить проверку существования пользователя с полученным guid в бд иначе 404
		if err != nil {
			ctx.String(http.StatusBadRequest, err.Error())
			return
		}
		// TODO: запись refreshToken в бд
		ctx.JSON(http.StatusOK, gin.H{ // TODO: запись в куки
			"access": accessToken,
			"refresh": refreshToken,
		})
	})
	router.POST("/refresh", func(ctx *gin.Context) {
		type Body struct { // TODO: чтение из куки
			Access string `json:"access"`
			Refresh string `json:"refresh"`
		}
		body := Body{}
		err := ctx.BindJSON(&body)
		if err != nil {
			ctx.String(http.StatusBadRequest, err.Error())
			return
		}
		accessClaims, err := jwtManager.ValidateAccessToken(body.Access)
		if err != nil {
			ctx.String(http.StatusBadRequest, err.Error())
			return
		}
		refreshClims, err := jwtManager.ValidateRefreshToken(body.Refresh)
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
		accessToken, refreshToken, err := jwtManager.RefreshTokenPair(accessClaims, refreshClims)
		ctx.JSON(http.StatusOK, gin.H{ // TODO: запись в куки
			"access": accessToken,
			"refresh": refreshToken,
		})
	})
	router.Run(":8080") // TODO: адрес из конфиг файла
}