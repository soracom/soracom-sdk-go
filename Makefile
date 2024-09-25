.PHONY: setup
setup:
	go get .

.PHONY: lint
lint:
	staticcheck ./...

.PHONY: test
test:
	go vet ./...
	go test -v -race -p=1 ./...

.PHONY: fmt
fmt:
	gofmt -w -s .
	goimports -w .

.PHONY: fmt-check
fmt-check:
	gofmt -l -s .
	goimports -l .
