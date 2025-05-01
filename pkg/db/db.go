package db

import (
	"github.com/AlexandrShapkin/auth-go-test-task/pkg/models"
	"gorm.io/gorm"
)

func ConnectDB(dialector *gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(*dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.User{},
	)

	if err != nil {
		return nil, err
	}

	return db, nil
}