package main

import (
	"context"
	"github.com/fnaumov/stringsvc/pb"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

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
