BINARY := wtw

.PHONY: build run lint fmt clean e2e e2e-docker

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

e2e: build
	WTW=./$(BINARY) go test ./e2e/... -v -timeout 120s

e2e-docker:
	docker run --rm \
		-v "$(shell pwd):/workspace" \
		-w /workspace \
		golang:1.26 \
		sh -c "apt-get update -qq && apt-get install -y -qq rsync git > /dev/null && \
			git config --global user.email test@test.com && \
			git config --global user.name Test && \
			git config --global init.defaultBranch main && \
			go build -o wtw . && \
			WTW=/workspace/wtw go test ./e2e/... -v -timeout 120s"

clean:
	rm -f $(BINARY)
