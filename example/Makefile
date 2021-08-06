all: server cli

server: ../jotto/*.go protocol/*.go apps/server/*.go proto
	go build -o bin/server ./apps/server

cli: ../jotto/*.go protocol/*.go commands/*.go apps/cli/*.go proto
	go build -o bin/cli ./apps/cli

proto:
	protoc --gogofast_out=./ ./protocol/protocol.proto
