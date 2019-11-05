package main

import (
	"context"
	"encoding/json"
	"github.com/fnaumov/stringsvc/pb"
	"github.com/go-kit/kit/endpoint"
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

type HealthRequest struct {}

type HealthResponse struct {
	S bool `json:"status"`
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

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func decodeHealthRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return HealthRequest{}, nil
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
		return HealthResponse{S: status}, nil
	}
}

// GRPC Binding

type grpcBinding struct {
	svc StringService
}

func (g grpcBinding) Uppercase(ctx context.Context, req *pb.UppercaseRequest) (*pb.UppercaseResponse, error) {
	v, err := g.svc.Uppercase(req.S)
	return &pb.UppercaseResponse{V: v, Err: ""}, err
}

func (g grpcBinding) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	v := g.svc.Count(req.S)
	return &pb.CountResponse{V: int64(v)}, nil
}

func (g grpcBinding) HealthCheck(ctx context.Context, req *pb.CountRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
}
