package app

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

func EncodeTokenToBase64(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func DecodeTokenFromBase64(b64token string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64token)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func HashToken(b64token string) (string, error) {
	sha := sha256.Sum256([]byte(b64token))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CompareHashAndToken(hashedToken string, b64token string) bool {
	sha := sha256.Sum256([]byte(b64token))
	err := bcrypt.CompareHashAndPassword([]byte(hashedToken), sha[:])
	return err == nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CompareHashAndPassword(hashedPassword string, rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	return err == nil
}
