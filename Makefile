APP_NAME:=sunbeam

.PHONY: build-frontend
build-frontend:
	cd frontend && npm run build

.PHONY: build
build: build-frontend
	go build -o bin/$(APP_NAME) .

.PHONY: start
run: start
	./bin/$(APP_NAME)

.PHONY: start-server
start-server: build
	./bin/$(APP_NAME) serve
