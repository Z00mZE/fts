#!/usr/bin/env bash
echo "Upgrade go dependencies"
go get -u all
go mod download
go mod tidy

echo "Install go dependencies"
echo "Install protoc-gen-grpc-gateway"
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
echo "Install protoc-gen-go"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
echo "Install protoc-gen-go-grpc"
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
echo "Install protoc-gen-openapiv2"
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

declare -A deps
deps["google/api/annotations.proto"]="https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto"
deps["google/api/http.proto"]="https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto"
deps["google/type/date.proto"]="https://raw.githubusercontent.com/googleapis/googleapis/master/google/type/date.proto"

for file in "${!deps[@]}"; do
  src=$(dirname ${file})
  if [ ! -d "$src" ]; then
    mkdir -p $src
  fi
  echo "Download \"${file}\" from \"${deps[$file]}\""
  curl --silent ${deps[$file]} >$file
done

echo "generating protobuf message types"
protoc \
  --go_out=./ \
  --go_opt="module=github.com/Z00mze/fts" \
  ./proto/*.proto

echo "generating gRPC service"
protoc \
  --go-grpc_out=require_unimplemented_servers=false:. \
  --go-grpc_opt="module=github.com/Z00mze/fts" \
  ./proto/*.proto

echo "generating gRPC-API-Gateway service"
protoc \
  --grpc-gateway_out=. \
  --grpc-gateway_opt="module=github.com/Z00mze/fts" \
  --openapiv2_out ./ \
   ./proto/*.proto

go mod tidy
go vet ./...
