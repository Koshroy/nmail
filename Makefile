.PHONY: all test
all: test nncp

nncp: main.go
	go build

test: main.go main_test.go
	go test
