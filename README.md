# StringSVC

Simple microservice on **Go Kit**

- Support HTTP and GRPC protocols
- Support logging method calls
- Implemented registration of services in **Consul** and health check method
- Implemented authorization with JWT Token
- Possibility deploy to **Kubernetes**

## Consul
- Install on macOS: `brew install consul`
- Start on macOS: `consul agent -dev`
- Dashboard panel: `http://localhost:8500/ui/dc1/services`

## Deployment to Kubernetes
- Docker build and tagging (run in Dockerfile directory) `docker build -t fnaumov/stringsvc .`
- Deployment command `kubectl create -f ks-manifest.yaml`

## Usage microservice
- Run
- Authorization request and receiving JWT token
```shell script
curl -v -XPOST -d '{"username": "user1", "password": "passwordOne"}' http://localhost:8080/auth
```
- Request (need specify token returned from auth request)
```shell script
curl -v -XPOST -d '{"s": "Hello world!"}' -H "Authorization: Bearer eyJhbGciOi..." http://localhost:8080/uppercase
```
