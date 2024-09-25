generate:
	/usr/local/bin/ocb --config builder-config.yaml

release:
	cd ./cmd && \
	go mod tidy && \
	go mod vendor && \
	GO111MODULE=on CGO_ENABLED=0 gox

dev-release:
	cd ./cmd && \
	go mod tidy && \
	go mod vendor && \
	GO111MODULE=on CGO_ENABLED=0 go build -o ../bin/core-collector