.PHONY: install test build serve clean pack deploy ship

TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)

export TAG
export PORT=3000
export GCP_PROJECT=plori-nicholss
export GCP_PATH=us.gcr.io/${GCP_PROJECT}
#export BOLT_PATH=/etc/proxy
export BOLT_PATH=.

install:
	go get .

test:
	go test ./...

build: install
	go build -ldflags "-X main.version=$(TAG)" -o proxy .

serve: build cfg
	./proxy

clean:
	rm ./proxy
	rm ./config.yml

cfg:
	envsubst < ./config.yml.dist > ./config.yml

pack: build cfg
	GOOS=linux make build
	docker build -t ${GCP_PATH}/k8s-broker-proxy:$(TAG) .

run: pack
	docker run -d -p ${PORT}:${PORT} ${GCP_PATH}/k8s-broker-proxy:$(TAG)

stop:
	docker ps
	@read -p "--> $ docker stop " imageId; \
	docker stop $$imageId

upload: pack
	gcloud docker -- push ${GCP_PATH}/k8s-broker-proxy:$(TAG)

deploy:
	envsubst < k8s/deployment.yml | kubectl apply -f -

ship: test pack upload deploy clean
