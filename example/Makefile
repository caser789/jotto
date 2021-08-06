all: upper client

upper: *.go ../jotto/*.go protocol/*.go proto
	go build -o bin/upper main.go routes.go

client: *.go ../jotto/*.go protocol/*.go proto
	go build -o bin/client client.go

proto:
	protoc --gogofast_out=./ ./protocol/protocol.proto
