FROM golang:1.23-alpine
WORKDIR /project

RUN apk add --no-cache upx

COPY go.mod go.sum /project 
RUN go mod download

COPY cmd /project/cmd
COPY pkg /project/pkg
COPY main.go /project
RUN go build -ldflags '-s -w' -o /project/tinys3cli main.go

RUN upx /project/tinys3cli
RUN mv /project/tinys3cli /bin/tinys3cli