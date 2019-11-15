package main

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const expiration = 120

type AuthService interface {
	Auth(string, string) (string, error)
}

type authService struct {
	key     []byte
	clients map[string]string
}

type customClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func generateToken(signingKey []byte, username string) (string, error) {
	claims := customClaims{
		username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * expiration).Unix(),
			IssuedAt:  jwt.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

func (as authService) Auth(username string, password string) (string, error) {
	if as.clients[username] == password {
		signed, err := generateToken(as.key, username)
		if err != nil {
			return "", errors.New(err.Error())
		}
		return signed, nil
	}
	err := errors.New("incorrect credentials")
	return "", err
}
