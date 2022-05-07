build:
	gofumpt -w *.go

cover:
	go test -cover ./...

race:
	go test -race