# genproto.ps1 — regenerate all gRPC / protobuf code under genproto/ from protos/.
# PowerShell counterpart of genproto.sh (for Windows hosts without Git Bash).
#
# Requirements (all on PATH unless noted):
#   - protoc                    (>= 3.19; verified with 28.x)
#   - protoc-gen-go             go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#   - protoc-gen-go-grpc        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#   - python + grpcio-tools     pip install grpcio-tools
#   - grpc_tools_node_protoc    npm install -g grpc-tools
#   - protoc-gen-grpc-java      https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/
#   - protoc-gen-grpc_csharp    provided by the Grpc.Tools NuGet package (tools/windows_x64/)

$ErrorActionPreference = "Stop"
$root = $PSScriptRoot
Set-Location $root

$serviceProtos = @("demo","cart","checkout","product_catalog","shipping","currency","payment","email","recommendation","shopping_assistant","ad")
$allProtos = $serviceProtos + @("grpc/health/v1/health")
$goPrefix = "github.com/Ayanamiz41/microservice-demo/genproto/go"

New-Item -ItemType Directory -Force genproto/go, genproto/python, genproto/nodejs, genproto/java, genproto/csharp | Out-Null

function Need([string]$cmd) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
    Write-Error "missing required tool: $cmd"
  }
}

Write-Host "== Go -> genproto/go =="
Need protoc; Need protoc-gen-go; Need protoc-gen-go-grpc
& protoc -I protos `
  --go_out=genproto/go --go_opt=module=$goPrefix `
  --go-grpc_out=genproto/go --go-grpc_opt=module=$goPrefix `
  ($serviceProtos | ForEach-Object { "$_.proto" })
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "== Python -> genproto/python =="
python -c "import grpc_tools" 2>$null
if ($LASTEXITCODE -ne 0) { Write-Error "grpcio-tools not installed (pip install grpcio-tools)" }
& python -m grpc_tools.protoc -I protos `
  --python_out=genproto/python --grpc_python_out=genproto/python `
  ($serviceProtos | ForEach-Object { "$_.proto" })
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "== Node.js -> genproto/nodejs =="
$grpcToolsProtoc = Get-Command grpc_tools_node_protoc -ErrorAction SilentlyContinue
if (-not $grpcToolsProtoc) { Write-Error "missing required tool: grpc_tools_node_protoc (npm install -g grpc-tools)" }
$nodeArgs = @("-I", "protos", "--js_out=import_style=commonjs,binary:genproto/nodejs", "--grpc_out=grpc_js:genproto/nodejs") + ($allProtos | ForEach-Object { "$_.proto" })
# The npm shim is a .cmd/.bat wrapper; call the underlying JS directly to avoid
# PowerShell->cmd argument mangling of the comma inside --js_out.
$gtnpPath = $grpcToolsProtoc.Source
if ($gtnpPath -like '*.cmd' -or $gtnpPath -like '*.bat') {
  $nodeModules = Split-Path (Split-Path $gtnpPath -Parent) -Parent
  $gtnpPath = Join-Path $nodeModules 'grpc-tools/bin/grpc_tools_node_protoc.js'
}
if ($gtnpPath -like '*.js') {
  & node $gtnpPath @nodeArgs
} else {
  & $grpcToolsProtoc @nodeArgs
}
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "== Java -> genproto/java =="
Need protoc-gen-grpc-java
& protoc -I protos `
  --java_out=genproto/java --grpc-java_out=genproto/java `
  ($allProtos | ForEach-Object { "$_.proto" })
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "== C# -> genproto/csharp =="
Need protoc-gen-grpc_csharp
& protoc -I protos `
  --csharp_out=genproto/csharp --grpc_csharp_out=genproto/csharp `
  ($allProtos | ForEach-Object { "$_.proto" })
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "genproto: all languages generated."
