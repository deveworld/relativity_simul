format:
	go fmt .

check:
	golangci-lint run

test:
	go test .

run:
	go run .
