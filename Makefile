run:
	go run cmd/main.go

rpcGen:
	protoc --go_out=../GeneratedProto --go-grpc_out=../GeneratedProto  internal/proto/api.proto