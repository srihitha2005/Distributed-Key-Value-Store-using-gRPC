# Install protobuf compiler
sudo apt update
sudo apt install -y protobuf-compiler

# Install Go plugins for protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add Go bin to PATH (for this session)
export PATH="$PATH:$(go env GOPATH)/bin"

# Generate Go code from proto
protoc --go_out=. --go-grpc_out=. proto/kvstore.proto
