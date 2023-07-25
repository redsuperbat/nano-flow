FROM golang:latest AS builder
WORKDIR /app

# Install deps
COPY go.mod go.sum ./
RUN go mod download

# Install protobuf cli
RUN apt-get update
RUN apt install -y protobuf-compiler protoc-gen-go-grpc protoc-gen-go

# Build binary
COPY . .
RUN protoc --go_out=. --go-grpc_out=. proto/message.proto
RUN go build -o nano-flow


FROM debian
COPY --from=builder /app/nano-flow /
ENTRYPOINT ["/nano-flow"]