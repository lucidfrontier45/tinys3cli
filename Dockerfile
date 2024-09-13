FROM golang:1.23-alpine AS builder
WORKDIR /project

RUN apk add --no-cache upx

COPY go.mod go.sum /project 
RUN go mod download

COPY cmd /project/cmd
COPY pkg /project/pkg
COPY main.go /project
RUN go build -ldflags '-s -w' -o /project/tinys3cli main.go
RUN upx /project/tinys3cli

FROM alpine 
COPY --from=builder /project/tinys3cli /bin/tinys3cli