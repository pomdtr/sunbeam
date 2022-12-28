APP_NAME:=sunbeam

.PHONY: init-frontend
init-frontend:
	npm --prefix web/frontend install

.PHONY: init-gui
init-gui:
	npm --prefix gui install

.PHONY: init
init: init-frontend init-gui

.PHONY: build-frontend
build-frontend:
	npm --prefix web/frontend run build

.PHONY: build
build: build-frontend
	go build -o bin/$(APP_NAME) .

.PHONY: run
run: build
	./bin/$(APP_NAME)

.PHONY: install
install: build-frontend
	go install

PORT := 8000
.PHONY: serve
serve: install
	sunbeam serve -p $(PORT)

.PHONY: gui
gui: install
	npm --prefix gui run start

.PHONY: install-gui
install-gui: install
	npm --prefix gui run install
