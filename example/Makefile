all: server cli

server: *.go ../jotto/*.go protocol/*.go apps/server/*.go proto
	go build -o bin/server ./apps/server

cli: *.go ../jotto/*.go protocol/*.go commands/*.go apps/cli/*.go proto
	go build -o bin/cli ./apps/cli

client: *.go ../jotto/*.go protocol/*.go proto
	go build -o bin/client ./apps/client

proto:
	protoc --gogofast_out=./ ./protocol/protocol.proto
