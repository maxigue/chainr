NAME=notif
BIN_DIR=bin
HELM_TMP_DIR=deployments/helm/tmp

DOCKER_TAG=tyrame/chainr-$(NAME):latest
DOCKERFILE=build/package/Dockerfile

.PHONY: all build run test fmt check-fmt docker-build deploy clean

all: build

build:
	go build -o $(BIN_DIR)/$(NAME)

run: build
	$(BIN_DIR)/$(NAME)

test:
	go test ./...

fmt:
	go fmt ./...

check-fmt:
	../scripts/check-fmt.sh

docker-build:
	docker build -t $(DOCKER_TAG) -f $(DOCKERFILE) .

deploy:
	./scripts/deploy.sh

clean:
	go clean
	rm -rf $(BIN_DIR)
	rm -rf $(HELM_TMP_DIR)
