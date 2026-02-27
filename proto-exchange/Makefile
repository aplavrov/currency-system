# USING protoc
.PHONY: protoc-generate
protoc-generate:
	protoc -I . exchange/exchange.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative
