
.PHONY=run
run:
	@echo ">> run: "
	go run -mod=vendor main.go -logtostderr

.PHONY=build-win
build-win:
	@echo ">> build windows program:"
	GOOS=windows go build -mod=vendor  -o bin/fix-network-win main.go
