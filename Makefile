
setup:
	go get .

test:
	go vet ./...
	go test -v -race -p=1 ./...
