
setup:
	go get .

lint:
	golint ./...

test:
	go vet ./...
	go test -v -race -p=1 ./...
