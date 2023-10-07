BINARY_NAME=blang

all: build test

build:
	go build -o ${BINARY_NAME} .

test:
	go test -v ./...

run:
	go run . -o example example.bl

clean:
	go clean
	rm test
	rm ${BINARY_NAME}