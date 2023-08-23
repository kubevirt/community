IMAGE_NAME=golang:1.18

generate:
	go run ./tools/sigs/generator/app.go

generate-dockerized:
	docker run --rm -e WHAT -e GOPROXY -v $(shell pwd):/go/src/app:Z $(IMAGE_NAME) make -C /go/src/app generate
