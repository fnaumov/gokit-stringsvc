package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fnaumov/stringsvc/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net/http"
	"testing"
)

var (
	consulClient = ConsulClient(consulAddr)
	svc StringService
)

func TestHTTPServer(t *testing.T) {
	runHTTPServer(consulClient, makeSvc(), httpAddr)
	jwtToken := httpJwtAuth(t)
	httpUppercase(t, jwtToken)
}

func httpJwtAuth(t *testing.T) string {
	requestBody, _ := json.Marshal(authRequest{
		Username: "user1",
		Password: "passwordOne",
	})

	resp, err := http.Post("http://localhost:8080/auth", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	var response authResponse
	_ = json.NewDecoder(resp.Body).Decode(&response)

	fmt.Println(fmt.Sprintf("response: %s", response.Token))

	return response.Token
}

func httpUppercase(t *testing.T, jwtToken string) {
	requestBody, _ := json.Marshal(uppercaseRequest{
		S: "Hello, this response for HTTP request!",
	})

	req, err := http.NewRequest("POST", "http://localhost:8080/uppercase", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	var response uppercaseResponse
	_ = json.NewDecoder(resp.Body).Decode(&response)

	fmt.Println(fmt.Sprintf("response: %s", response.V))
	assert.Equal(t, response.V, "HELLO, THIS RESPONSE FOR HTTP REQUEST!")
}

func TestGRPCServer(t *testing.T) {
	runGRPCServer(consulClient, makeSvc(), grpcAddr)

	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	client := pb.NewStringServiceClient(conn)
	request := &pb.UppercaseRequest{
		S: "Hello, this response for GRPC request!",
	}

	response, err := client.Uppercase(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("response: %s", response.V))
	assert.Equal(t, response.V, "HELLO, THIS RESPONSE FOR GRPC REQUEST!")
}

func makeSvc() StringService {
	svc = stringService{authConfig}
	svc = loggingMiddleware{authConfig, logger, svc}
	return svc
}