package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AlexandrShapkin/auth-go-test-task/app"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/config"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/db"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/jwt"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/mailer"
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Функция обязана загрузить конфигурацию, иначе продолжать работу программы нет смысла
func mustLoadConfig() *config.Config {
	cfg, err := config.LoadConfig("config", "yaml", ".")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1) // насколько знаю в таких случаях используется panic(), для того чтобы ее можно было отловить, но не думаю что здесь это требуется
	}
	return cfg
}

// Функция обязана совершить успешное подключение к базе данных, иначе продолжать работу программы нет смысла
func mustConnectDB(dbCfg config.Database) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbCfg.Host,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.DBName,
		dbCfg.Port,
		dbCfg.SSLMode,
	)
	dialector := postgres.Open(dsn)
	database, err := db.ConnectDB(&dialector)
	if err != nil {
		slog.Error("Failed to connet database", "error", err)
		os.Exit(1) // аналогично с комментарием оставленным в mustLoadConfig
	}
	return database
}

func main() {
	cfg := mustLoadConfig()

	jwtManager := jwt.NewJWT(
		[]byte(cfg.JWT.AccessSecretKey),
		[]byte(cfg.JWT.RefreshSecretKey),
		jwt.AccessExpires,
		jwt.RefreshExpires,
		jwt.ParseLeewayWindow,
	)

	database := mustConnectDB(cfg.Database)
	userRepo := repositories.NewUserRepo(database)

	mailer := mailer.NewMailer(cfg.Mail.From, cfg.Mail.Pass)

	application := app.NewApp(
		jwtManager,
		userRepo,
		mailer,
		cfg.App.LoginRemoteIPMode,
		cfg.App.RefreshRemoteIPMode,
		cfg.App.Domain,
	)
	application.Run(cfg.App.Addr)
}
