VERSION=1.0
BINARY=webbench
linux:
	@echo "build for linux..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "-s -w -X main.version=${VERSION}" -o ${BINARY}
windows:
	@echo "build for windows"
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GO111MODULE=on go build -ldflags "-s -w -X main.version=${VERSION}" -o ${BINARY}.exe
clean:
	rm $(BINARY)
	go clean
install:
	@echo "install dependencies"
	GO111MODULE=on go install