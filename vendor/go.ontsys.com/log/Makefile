.PHONY: lint prep test watch viewcover vendor

lint:
	go get -u golang.org/x/tools/cmd/vet
	go get -u github.com/golang/lint/golint
	golint ./... | grep -v '^vendor\/' | sed 's/^/golint: /' || true
	go vet ./... 2>&1 | grep -v '^vendor\/' | grep -v '^exit\ status\ 1' | sed 's/^/go vet: /' || true

prep:
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega

test: prep
	mkdir -p test_results
	GO15VENDOREXPERIMENT=1 ginkgo -r -cover

watch:
	GO15VENDOREXPERIMENT=1 ginkgo watch -r -cover

cover:
	go tool cover -html=log.coverprofile
	go tool cover -html=awslog/awslog.coverprofile

vendor:
	go get github.com/tools/godep
	GO15VENDOREXPERIMENT=1 godep save ./...
