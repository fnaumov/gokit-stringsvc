package main

import (
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	gokitjwt "github.com/go-kit/kit/auth/jwt"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
)

// Decoders and Encoders

func decodeUppercaseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	return request, nil
}

func decodeCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request countRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	return request, nil
}

func decodeHealthRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return healthRequest{}, nil
}

func decodeAuthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request authRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// HTTP Handler

func makeHTTPHandler(svc StringService) http.Handler {
	kf := func(token *jwt.Token) (interface{}, error) {
		return authConfig.key, nil
	}
	clf := func() jwt.Claims {
		return &customClaims{}
	}
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(gokitjwt.HTTPToContext()),
	}

	r := mux.NewRouter()

	r.Methods("POST").Path("/uppercase").Handler(httptransport.NewServer(
		gokitjwt.NewParser(kf, jwt.SigningMethodHS256, clf)(makeUppercaseEndpoint(svc)),
		decodeUppercaseRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/count").Handler(httptransport.NewServer(
		gokitjwt.NewParser(kf, jwt.SigningMethodHS256, clf)(makeCountEndpoint(svc)),
		decodeCountRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		makeHealthEndpoint(svc),
		decodeHealthRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/auth").Handler(httptransport.NewServer(
		makeAuthEndpoint(svc),
		decodeAuthRequest,
		encodeResponse,
		options...,
	))

	return r
}
