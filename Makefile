gen-code:
	buf generate
build:
	go build
run:
	go build && ./grpc-lb
