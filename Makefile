# Go パラメータ
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=AnswerButton
BINARY_WIN=$(BINARY_NAME).exe
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build
build:
	$(GOBUILD) -o $(BINARY_WIN) -v
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	$(GOGET) github.com/markbates/goth
	$(GOGET) github.com/markbates/pop


# クロスコンパイル
build-win:
	set CGO_ENABLED=1&&set GOOS=windows&& set GOARCH=amd64&& $(GOBUILD) -o $(BINARY_WIN) -v -ldflags -H=windowsgui

build-mac:
	set CGO_ENABLED=1&&set GOOS=darwin&& set GOARCH=amd64&& $(GOBUILD) -o $(BINARY_NAME) -v

build-linux:
	set CGO_ENABLED=1&& set GOOS=linux&& set GOARCH=amd64&& $(GOBUILD) -o $(BINARY_UNIX) -v

docker-build:
	docker run --rm -it -v "$(GOPATH)":/go -w /go/src/bitbucket.org/rsohlich/makepost golang:latest go build -o "$(BINARY_UNIX)" -v
