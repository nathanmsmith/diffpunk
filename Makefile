.PHONY: frontend build lint-go test browser-test release clean

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
	go test -race ./...
	npm run test:dom
	npm run test:cli

browser-test: build
	npx playwright install --with-deps chromium
	npm run test:browser
	npm run test:dev

release: frontend
	./scripts/release.sh

clean:
	rm -rf bin release cli/assets/dist
	mkdir -p cli/assets/dist
	touch cli/assets/dist/.gitkeep
