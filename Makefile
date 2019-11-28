
setup:
	go get .

test:
	go vet ./...

	# test api client (ignore metadata client)
	go test -v -race -p=1 ./... -run "^Test[ABCDEFGHIJKLNOPQRSTUVWXYZ].+"
