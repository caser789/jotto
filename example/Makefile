all: server cli worker

server: ../motto/*.go protocol/*.go apps/server/*.go proto
	go build -o bin/server ./apps/server

cli: ../motto/*.go protocol/*.go commands/*.go apps/cli/*.go proto
	go build -o bin/cli ./apps/cli

worker: ../motto/*.go protocol/*.go commands/*.go apps/worker/*.go proto
	go build -o bin/worker ./apps/worker

proto:
	protoc --gogofast_out=./ ./protocol/protocol.proto
