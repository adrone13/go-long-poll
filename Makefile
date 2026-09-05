.PHONY: run-server run-client test-race

run-server:
	go run ./server

run-client:
	go run ./client

test-race:
	go test -race ./server/... -v