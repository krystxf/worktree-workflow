BINARY := wtw

.PHONY: build run lint fmt clean

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

clean:
	rm -f $(BINARY)
