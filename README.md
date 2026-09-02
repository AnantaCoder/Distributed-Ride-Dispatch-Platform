# Distributed-Ride-Dispatch-Platform

## Protobuf / Code Generation
This project uses `buf` to manage and generate Protocol Buffers.

### 1. Install Dependencies
You must have Go installed. Then, install `buf` and the necessary Go plugins:
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
```

### 2. Generate Code
To generate the Go code and ConnectRPC stubs from the `.proto` files, run:
```bash
cd proto
buf generate
```
This will output all the generated code into the `gen/` directory.

### 3. Linting
To check your `.proto` files against the standard Protobuf style guide, run:
```bash
cd proto
buf lint
```

## References
* [Remote Procedure Calls, Protocol Buffers, and Modern Distributed Systems Communication](https://www.freecodecamp.org/news/remote-procedure-calls-protocol-buffers-and-modern-distributed-systems-communication/)