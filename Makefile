# Image URL to use all building/pushing image targets
IMG ?= quay.io/t3n/krelabel:dev

# Build binary
bin: fmt vet
	CGO_ENABLED=0 go build -a -o build/krelabel main.go

# Run kacheproxy
run: fmt vet
	go run main.go

# Run go fmt against code
fmt:
	go fmt

# Run go vet against code
vet:
	go vet

# Build the docker image
docker-build:
	docker build . -t ${IMG} -f Dockerfile

# Push the docker image
docker-push:
	docker push ${IMG}
