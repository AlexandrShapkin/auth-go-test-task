package main

import (
	"fmt"
	"log/slog"

	"github.com/AlexandrShapkin/auth-go-test-task/app"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/config"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/db"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/jwt"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/repositories"
	"gorm.io/driver/postgres"
)

func main() {
	cfg, err := config.LoadConfig("config", "yaml", ".")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
	}

	jwtManager := jwt.NewJWT(
		[]byte(cfg.JWT.AccessSecretKey),
		[]byte(cfg.JWT.RefreshSecretKey),
		jwt.AccessExpires,
		jwt.RefreshExpires,
		jwt.ParseLeewayWindow,
	)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)
	dialector := postgres.Open(dsn)
	database, err := db.ConnectDB(&dialector)
	if err != nil {
		panic(err)
	}
	userRepo := repositories.NewUserRepo(database)

	application := app.NewApp(jwtManager, userRepo)
	application.Run(":8080")
}
