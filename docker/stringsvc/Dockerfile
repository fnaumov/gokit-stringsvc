FROM golang:1.13-alpine as builder
RUN apk add --no-cache ca-certificates git

ENV PROJECT github.com/fnaumov/stringsvc
WORKDIR /go/src

RUN go get $PROJECT
WORKDIR ./$PROJECT

RUN go build -o /go/bin/stringsvc

FROM alpine as release
RUN apk add --no-cache ca-certificates \
    busybox-extras net-tools bind-tools

COPY --from=builder /go/bin/stringsvc /stringsvc

ENTRYPOINT ["/stringsvc"]
