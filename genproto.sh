#!/usr/bin/env bash
# genproto.sh — regenerate all gRPC / protobuf code under genproto/ from protos/.
#
# Requirements (all on PATH unless noted):
#   - protoc                    (>= 3.19; verified with 28.x)  e.g. `choco install protobuf`
#   - protoc-gen-go              go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#   - protoc-gen-go-grpc         go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#   - python + grpcio-tools      pip install grpcio-tools       (used via `python -m grpc_tools.protoc`)
#   - grpc_tools_node_protoc     npm install -g grpc-tools      (Node.js codegen)
#   - protoc-gen-grpc-java       https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/
#   - protoc-gen-grpc_csharp     provided by the Grpc.Tools NuGet package (tools/windows_x64/)
#
# Exclusions (intentional):
#   - Go:     grpc/health/v1/health.proto is skipped — Go services use the
#             canonical google.golang.org/grpc/health/grpc_health_v1 package.
#   - Python: grpc/health/v1 is skipped — a local grpc/ dir would shadow the
#             grpc runtime package on sys.path; use grpcio-health-checking
#             (top-level grpc_health.v1) instead.
#   Node/Java/C# include the health contract in their generated trees.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

PROTOC="${PROTOC:-protoc}"
PYTHON="${PYTHON:-python3}"

GO_PREFIX="github.com/Ayanamiz41/microservice-demo/genproto/go"

mkdir -p genproto/go genproto/python genproto/nodejs genproto/java genproto/csharp

SERVICE_PROTOS=(demo cart checkout product_catalog shipping currency payment email recommendation shopping_assistant ad)
ALL_PROTOS=("${SERVICE_PROTOS[@]}" grpc/health/v1/health)

need() { command -v "$1" >/dev/null 2>&1 || { echo "ERROR: missing required tool: $1" >&2; exit 1; }; }

echo "== Go -> genproto/go =="
need "$PROTOC"; need protoc-gen-go; need protoc-gen-go-grpc
"$PROTOC" -I protos \
  --go_out=genproto/go --go_opt=module="$GO_PREFIX" \
  --go-grpc_out=genproto/go --go-grpc_opt=module="$GO_PREFIX" \
  "${SERVICE_PROTOS[@]/%/.proto}"

echo "== Python -> genproto/python =="
"$PYTHON" -c "import grpc_tools" 2>/dev/null || { echo "ERROR: grpcio-tools not installed (pip install grpcio-tools)" >&2; exit 1; }
"$PYTHON" -m grpc_tools.protoc -I protos \
  --python_out=genproto/python --grpc_python_out=genproto/python \
  "${SERVICE_PROTOS[@]/%/.proto}"

echo "== Node.js -> genproto/nodejs =="
need grpc_tools_node_protoc
grpc_tools_node_protoc -I protos \
  --js_out=import_style=commonjs,binary:genproto/nodejs \
  --grpc_out=grpc_js:genproto/nodejs \
  "${ALL_PROTOS[@]/%/.proto}"

echo "== Java -> genproto/java =="
need protoc-gen-grpc-java
"$PROTOC" -I protos \
  --java_out=genproto/java --grpc-java_out=genproto/java \
  "${ALL_PROTOS[@]/%/.proto}"

echo "== C# -> genproto/csharp =="
need protoc-gen-grpc_csharp
"$PROTOC" -I protos \
  --csharp_out=genproto/csharp --grpc_csharp_out=genproto/csharp \
  "${ALL_PROTOS[@]/%/.proto}"

echo "genproto: all languages generated."
