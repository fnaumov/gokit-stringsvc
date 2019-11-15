package main

import (
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/fnaumov/stringsvc/pb"
	gokitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"net/http"
)

// Requests and Responses

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V string `json:"v"`
	Err string `json:"err,omitempty"`
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}

type healthRequest struct {}

type healthResponse struct {
	S bool `json:"status"`
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token,omitempty"`
	Err   string `json:"err,omitempty"`
}

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

// Endpoints

func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		v, err := svc.Uppercase(req.S)
		if err != nil {
			return uppercaseResponse{v, err.Error()}, nil
		}

		return uppercaseResponse{v, ""}, nil
	}
}

func makeCountEndpoint(svc StringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		v := svc.Count(req.S)

		return countResponse{v}, nil
	}
}

func makeHealthEndpoint(svc StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		status := svc.HealthCheck()
		return healthResponse{S: status}, nil
	}
}

func makeAuthEndpoint(svc StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(authRequest)
		token, err := svc.Auth(req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return authResponse{token, ""}, nil
	}
}

func makeHttpHandler(svc StringService) http.Handler {
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

// GRPC Binding

type grpcBinding struct {
	svc StringService
	healthServer *health.Server
}

func (g grpcBinding) Uppercase(ctx context.Context, req *pb.UppercaseRequest) (*pb.UppercaseResponse, error) {
	v, err := g.svc.Uppercase(req.S)
	return &pb.UppercaseResponse{V: v, Err: ""}, err
}

func (g grpcBinding) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	v := g.svc.Count(req.S)
	return &pb.CountResponse{V: int64(v)}, nil
}

func (g grpcBinding) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	res, err :=  g.healthServer.Check(ctx, req)
	return res, err
}

func (g grpcBinding) Watch(req *healthpb.HealthCheckRequest, hws healthpb.Health_WatchServer) error {
	err :=  g.healthServer.Watch(req, hws)
	return err
}

func (g grpcBinding) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	token, err := g.svc.Auth(req.Username, req.Password)
	return &pb.AuthResponse{Token: token, Err: ""}, err
}
