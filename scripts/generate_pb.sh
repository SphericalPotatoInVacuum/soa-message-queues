#!/usr/bin/env bash

DIR="$(pwd)"
mkdir -p proto_gen

# Find all directories containing at least one prototfile.
# Based on: https://buf.build/docs/migration-prototool#prototool-generate.
dirs="$( find ${DIR}/proto -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq )"
for dir in ${dirs}; do
  files=$(find "${dir}" -name '*.proto')

  # Generate all files with protoc-gen-go.
  protoc -I=${DIR}/proto \
    --go_out=${DIR}/proto_gen --go_opt=paths=source_relative \
    --go-grpc_out=${DIR}/proto_gen --go-grpc_opt=paths=source_relative \
    --experimental_allow_proto3_optional ${files}
done
