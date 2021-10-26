.PHONY: all test
all: test nncp

nncp: main.go send.go recv.go util.go
	go build

test: send_test.go recv_test.go util_test.go
	go test
