all: server cli worker

server: ../motto/*.go protocol/*.go apps/server/*.go proto
	go build -o bin/server ./apps/server

spex: ../motto/*.go protocol/*.go apps/spex/*.go proto
	go build -o bin/spex ./apps/spex

cli: ../motto/*.go protocol/*.go commands/*.go apps/cli/*.go proto
	go build -o bin/cli ./apps/cli

worker: ../motto/*.go protocol/*.go commands/*.go apps/worker/*.go proto
	go build -o bin/worker ./apps/worker

proto:
	protoc --gogofast_out=./ ./protocol/protocol.proto
