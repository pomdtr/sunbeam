APP_NAME:=sunbeam

.PHONY:
init:
	cd web/frontend && npm install

.PHONY: build-frontend
build-frontend:
	cd web/frontend && npm run build

.PHONY: build
build: build-frontend
	go build -o bin/$(APP_NAME) .

.PHONY: run
run: build
	./bin/$(APP_NAME)

.PHONY: install
install: build-frontend
	go install

.PHONY: serve
serve: build
	./bin/$(APP_NAME) serve
