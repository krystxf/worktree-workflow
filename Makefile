BINARY := wtw

.PHONY: build run lint fmt clean e2e

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

e2e: build
	WTW=./$(BINARY) ./e2e/run.sh

clean:
	rm -f $(BINARY)
