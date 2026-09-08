#!/bin/bash

# Exit immediately if any command fails
set -e

echo "🔄 Compiling Protobuf profiles from root..."


mkdir -p ./vital_gateway/proto/payment_pb
mkdir -p ./payment_service/proto/pb


echo "🐹 Generating Go stubs for vital_gateway..."

protoc --proto_path=./a_proto \
       --go_out=./vital_gateway \
       --go-grpc_out=./vital_gateway \
       ./a_proto/payment.proto


echo "⚡ Generating JavaScript stubs via npx..."
npx -p grpc-tools -p ts-protoc-gen grpc_tools_node_protoc \
    --js_out=import_style=commonjs,binary:./payment_service/proto/pb \
    --grpc_out=grpc_js:./payment_service/proto/pb \
    --plugin=protoc-gen-ts=./payment_service/node_modules/.bin/protoc-gen-ts \
    --ts_out=grpc_js:./payment_service/proto/pb \
    -I=a_proto \
    a_proto/payment.proto

echo "✅ Success! Go stubs generated in vital_gateway and JS stubs generated in payment_service."