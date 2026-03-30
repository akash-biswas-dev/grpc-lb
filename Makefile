gen-code:
	buf generate

build:
	go build && ./grpc-lb
