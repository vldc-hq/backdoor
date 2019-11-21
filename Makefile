ci: main.go
	go build -o ci .

lint:
	golangci-lint run

fmt:
	gofmt -w -s .
	goimports -w .
