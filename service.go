package main

import (
	"errors"
	"strings"
)

type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
	HealthCheck() bool
	Auth(string, string) (string, error)
}

type stringService struct {
	auth AuthService
}

var ErrEmpty = errors.New("empty string")

func (ss stringService) Uppercase(s string) (token string, err error) {
	if s == "" {
		return "", ErrEmpty
	}

	return strings.ToUpper(s), nil
}

func (ss stringService) Count(s string) int {
	return len(s)
}

func (ss stringService) HealthCheck() bool {
	return true
}

func (ss stringService) Auth(username string, password string) (token string, err error) {
	token, err = ss.auth.Auth(username, password)
	return token, err
}
