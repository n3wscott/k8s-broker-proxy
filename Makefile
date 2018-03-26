TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)

SOURCES       := $(shell find . -name '*.go' -not -path "*/vendor/*")
SOURCE_DIRS    = pkg
.DEFAULT_GOAL := check

export TAG
export PORT=3000
export GCP_PROJECT=plori-nicholss
export GCP_PATH=us.gcr.io/${GCP_PROJECT}

install: ## Go get deps
	@go get .

test: ## Run unit tests
	@go test -cover ./pkg/...

build: install ## Build the proxy output
	@go build -ldflags "-X main.version=$(TAG)" -o proxy .

fmtcheck: ## Check go formatting
	@gofmt -l $(SOURCES) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

serve: build cfg ## Run the proxy locally
	./proxy

clean: ## Clean
	rm ./proxy
	rm ./config.yml

cfg: ## Write the config into envsubst
	envsubst < ./config.yml.dist > ./config.yml

check: fmtcheck vet lint test ## Pre-flight checks before creating PR

lint: ## Run golint
	@golint -set_exit_status $(addsuffix /... , $(SOURCE_DIRS))

pack: build cfg ## Make a docker pack
	GOOS=linux make build
	docker build -t ${GCP_PATH}/k8s-broker-proxy:$(TAG) .

run: pack ## Run in docker
	docker run -d -p ${PORT}:${PORT} ${GCP_PATH}/k8s-broker-proxy:$(TAG)

stop: ## Stop docker
	docker ps
	@read -p "--> $ docker stop " imageId; \
	docker stop $$imageId

upload: pack ## Upload docker image to gcp
	gcloud docker -- push ${GCP_PATH}/k8s-broker-proxy:$(TAG)

deploy: ## kubectl apply the deployment to k8s
	envsubst < k8s/deployment.yml | kubectl apply -f -

ship: test pack upload deploy clean ## Test, pack, upload, deploy and then clean

vet: ## Run go vet
	@go tool vet ./pkg

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: install test build serve clean pack deploy ship vet cfg check fmtcheck