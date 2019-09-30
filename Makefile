.PHONY: lintinstall lint build prep test watch viewcover
.DEFAULT_GOAL := test

# Not a prerequisite of lint becuase it takes a while
lintinstall:
	@go get -u github.com/alecthomas/gometalinter
	@gometalinter --install --no-vendored-linters

lint:
	@gometalinter --vendor ./...

prep:
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega
	go get ./...

build:
	go build

test:
	mkdir -p test_results
	ginkgo -r -cover

watch:
	ginkgo watch -r -cover

cover:
	go tool cover -html=isclib.coverprofile
