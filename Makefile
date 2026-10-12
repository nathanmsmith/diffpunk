.PHONY: frontend build lint-go test release clean

frontend:
	npm ci
	npm run build

build: frontend
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o bin/dp ./cli

lint-go: frontend
	golangci-lint config verify
	golangci-lint run ./...

test: build lint-go
	npm run lint
	npm run format:check
	npm run typecheck
	go test -race -covermode=atomic -coverpkg=./cli/... -coverprofile=coverage.out ./cli/...

release: frontend
	./scripts/release.sh

clean:
	rm -rf bin release cli/assets/dist
	mkdir -p cli/assets/dist
	touch cli/assets/dist/.gitkeep
