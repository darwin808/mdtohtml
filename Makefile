EXECUTABLE_NAME := app

.PHONY: build
build: 
	@CGO_ENABLED=0 go build -o $(EXECUTABLE_NAME) .


.PHONY: run
run: 
	go run main.go
	