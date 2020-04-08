
setup:
	go get .

lint:
	golint ./...

test:
	go vet ./...
	go test -v -race -p=1 ./...

fmt:
	gofmt -w -s .

fmt-check:
	gofmt -l -s . | grep [^*][.]go$$; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -eq 0 ]; then exit 1; fi
