.PHONY: lintinstall lint build prep test watch viewcover
.DEFAULT_GOAL := test

# Not a prerequisite of lint because it takes a while
lintinstall:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@

lint:
	@golangci-lint run

prep:
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	go get github.com/onsi/gomega
	go get ./...

build:
	go build

test:
	mkdir -p test_results
	ginkgo -r -cover --junit-report test_results/junit-isclib.xml

watch:
	ginkgo watch -r -cover

cover:
	go tool cover -html=isclib.coverprofile
