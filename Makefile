
VER=$(shell git rev-parse --short HEAD)
d=$(shell date "+%Y-%m-%dT%H:%M:%S")

.PHONY=run
run:
	@echo ">> run: "
	go run -mod=vendor main.go -logtostderr

.PHONY=build-win
build-win:
	@echo ">> build windows program:"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build --ldflags="-X main.Version=$(VER) -X main.Date=$d" -mod=vendor  -o bin/fix-network.exe main.go

.PHONY=build
build:
	@echo ">> build program:"
	go build --ldflags "-X main.Version=$(VER) -X main.Date=$d" -mod=vendor  -o bin/fix-network main.go
