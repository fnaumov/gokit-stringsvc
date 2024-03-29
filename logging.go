package main

import (
	"github.com/go-kit/kit/log"
	"time"
)

type loggingMiddleware struct {
	auth AuthService
	logger log.Logger
	next StringService
}

func (mw loggingMiddleware) Uppercase(s string) (output string, err error) {
	defer func (begin time.Time) {
		_ = mw.logger.Log(
			"method", "uppercase",
			"input", s,
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.next.Uppercase(s)
	return
}

func (mw loggingMiddleware) Count(s string) (n int64) {
	defer func (begin time.Time) {
		_ = mw.logger.Log(
			"method", "count",
			"input", s,
			"n", n,
			"took", time.Since(begin),
		)
	}(time.Now())

	n = mw.next.Count(s)
	return
}

func (mw loggingMiddleware) HealthCheck() (n bool) {
	defer func (begin time.Time) {
		_ = mw.logger.Log(
			"method", "healthCheck",
			"n", n,
			"took", time.Since(begin),
		)
	}(time.Now())

	n = mw.next.HealthCheck()
	return
}

func (mw loggingMiddleware) Auth(clientID string, clientSecret string) (token string, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "auth",
			"username", clientID,
			"token", token,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	token, err = mw.auth.Auth(clientID, clientSecret)
	return
}
