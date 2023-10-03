BINARY_NAME=blang

all: build test

build:
	go build -o ${BINARY_NAME} .

test:
	go test ./tokeniser ./parser ./generator .

run:
	go run . test.bl

clean:
	go clean
	rm test
	rm ${BINARY_NAME}