package app

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// Кодирует сырой токен в строку base64
func EncodeTokenToBase64(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// Декодирует base64 токен обратно в сырой
func DecodeTokenFromBase64(b64token string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64token)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Хеширует base64 токен с использованием bcrypt
//
// !! Для того чтобы было возможно уместить токен в 72 бита он также будет захеширован при помощи sha256 !!
func HashToken(b64token string) (string, error) {
	sha := sha256.Sum256([]byte(b64token))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Сравнивает хешированный и base64 токены
//
// !! Передвать нужно именно base64 токен в качестве второго аргумента, не хеш sha256 и не сырой токен !!
func CompareHashAndToken(hashedToken string, b64token string) bool {
	sha := sha256.Sum256([]byte(b64token))
	err := bcrypt.CompareHashAndPassword([]byte(hashedToken), sha[:])
	return err == nil
}

// Хеширует пароль с использованием bcrypt для записи в БД
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Сравнивает хешированный bcrypt пароль и сырой пароль
func CompareHashAndPassword(hashedPassword string, rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	return err == nil
}
