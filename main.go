package main

import (
	"github.com/AlexandrShapkin/auth-go-test-task/app"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/jwt"
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
	
	application := app.NewApp(jwtManager)
	application.Run(":8080")
}