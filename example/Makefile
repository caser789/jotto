swap: *.go ../jotto/*.go protocol/*.go proto
	go build -o bin/upper .

proto:
	protoc --gogofast_out=./ ./protocol/protocol.proto
