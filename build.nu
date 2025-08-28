$env.CGO_ENABLED = "0"

GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o tinys3cli-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -ldflags "-w -s" -o tinys3cli-linux-arm64 .
GOOS=darwin GOARCH=arm64 go build -ldflags "-w -s" -o tinys3cli-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -ldflags "-w -s" -o tinys3cli-windows-amd64.exe .