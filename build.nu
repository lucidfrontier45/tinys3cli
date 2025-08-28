$env.CGO_ENABLED = "0"

GOOS=windows GOARCH=amd64 go build -ldflags "-w -s" -o tinys3cli-windows-amd64.exe .
upx tinys3cli-windows-amd64.exe

GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o tinys3cli-linux-amd64 .
upx tinys3cli-linux-amd64

GOOS=linux GOARCH=arm64 go build -ldflags "-w -s" -o tinys3cli-linux-arm64 .
upx tinys3cli-linux-arm64

GOOS=darwin GOARCH=arm64 go build -ldflags "-w -s" -o tinys3cli-darwin-arm64 .