FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
WORKDIR /project

RUN apk add --no-cache upx

COPY go.mod go.sum /project/
RUN go mod download

COPY cmd /project/cmd
COPY pkg /project/pkg
COPY main.go /project/

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '-s -w' -o /project/tinys3cli main.go
RUN upx /project/tinys3cli

FROM alpine 
COPY --from=builder /project/tinys3cli /tinys3cli