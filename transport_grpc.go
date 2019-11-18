package main

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/fnaumov/stringsvc/pb"
	gokitjwt "github.com/go-kit/kit/auth/jwt"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Decoders and Encoders

func decodeUppercaseGRPCRequest(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*pb.UppercaseRequest)
	return uppercaseRequest{S: r.S}, nil
}

func encodeUppercaseGRPCResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	r := resp.(uppercaseResponse)
	return &pb.UppercaseResponse{V: r.V, Err: r.Err}, nil
}

func decodeCountGRPCRequest(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*pb.CountRequest)
	return countRequest{S: r.S}, nil
}

func encodeCountGRPCResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	r := resp.(countResponse)
	return &pb.CountResponse{V: r.V}, nil
}

func decodeAuthGRPCRequest(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*pb.AuthRequest)
	return authRequest{Username: r.Username, Password: r.Password}, nil
}

func encodeAuthGRPCResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	r := resp.(authResponse)
	return &pb.AuthResponse{Token: r.Token, Err: r.Err}, nil
}

// GRPC Binding

type grpcBinding struct {
	svc StringService
	healthServer *health.Server
	uppercase grpctransport.Handler
	count grpctransport.Handler
	auth grpctransport.Handler
}

func (g grpcBinding) Uppercase(ctx context.Context, req *pb.UppercaseRequest) (*pb.UppercaseResponse, error) {
	_, response, err := g.uppercase.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.(*pb.UppercaseResponse), nil
}

func (g grpcBinding) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	_, response, err := g.count.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.(*pb.CountResponse), nil
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
	_, response, err := g.auth.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.(*pb.AuthResponse), nil
}

// GRPC Handler

func makeGRPCBinding(svc StringService, grpcBind grpcBinding) *grpcBinding {
	kf := func(token *jwt.Token) (interface{}, error) {
		return authConfig.key, nil
	}
	clf := func() jwt.Claims {
		return &customClaims{}
	}

	options := []grpctransport.ServerOption{
		grpctransport.ServerBefore(gokitjwt.GRPCToContext()),
	}

	grpcBind.uppercase = grpctransport.NewServer(
		gokitjwt.NewParser(kf, jwt.SigningMethodHS256, clf)(makeUppercaseEndpoint(svc)),
		decodeUppercaseGRPCRequest,
		encodeUppercaseGRPCResponse,
		options...,
	)

	grpcBind.count = grpctransport.NewServer(
		gokitjwt.NewParser(kf, jwt.SigningMethodHS256, clf)(makeCountEndpoint(svc)),
		decodeCountGRPCRequest,
		encodeCountGRPCResponse,
		options...,
	)

	grpcBind.auth = grpctransport.NewServer(
		makeAuthEndpoint(svc),
		decodeAuthGRPCRequest,
		encodeAuthGRPCResponse,
		options...,
	)

	return &grpcBind
}
