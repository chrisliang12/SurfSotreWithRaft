# Surfstore

This is the starter code for Project 4: Surfstore.

Before you get started, make sure you understand the following 2 things about Go. (These will also be covered in class and in discussions)
1. Interfaces: They are named collections of method signatures. Here are some good resources to understand interfaces in Go:
    a. https://gobyexample.com/interfaces
    b. https://jordanorelli.com/post/32665860244/how-to-use-interfaces-in-go

2. gRPC: You should know how to write gRPC servers and clients in Go. The [gRPC official documentation](https://grpc.io/docs/languages/go/basics/) of the *grpc* is a good resource.

## Protocol buffers

The starter code defines the following protocol buffer message type in `SurfStore.proto`:

```
message Block {
    bytes blockData = 1;
    int32 blockSize = 2;
}

message FileMetaData {
    string filename = 1;
    int32 version = 2;
    repeated string blockHashList = 3;
}
...
```

`SurfStore.proto` also defines the gRPC service:
```
service BlockStore {
    rpc GetBlock (BlockHash) returns (Block) {}
    rpc PutBlock (Block) returns (Success) {}
    rpc HasBlocks (BlockHashes) returns (BlockHashes) {}
}

service MetaStore {
    rpc GetFileInfoMap(google.protobuf.Empty) returns (FileInfoMap) {}
    rpc UpdateFile(FileMetaData) returns (Version) {}
    rpc GetBlockStoreAddr(google.protobuf.Empty) returns (BlockStoreAddr) {}
}
```

**You need to generate the gRPC client and server interfaces from our .proto service definition.** We do this using the protocol buffer compiler protoc with a special gRPC Go plugin (The [gRPC official documentation](https://grpc.io/docs/languages/go/basics/) introduces how to install the protocol compiler plugins for Go).

```shell
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/surfstore/SurfStore.proto
```

Running this command generates the following files in the `pkg/surfstore` directory:
- `SurfStore.pb.go`, which contains all the protocol buffer code to populate, serialize, and retrieve request and response message types.
- `SurfStore_grpc.pb.go`, which contains the following:
	- An interface type (or stub) for clients to call with the methods defined in the SurfStore service.
	- An interface type for servers to implement, also with the methods defined in the SurfStore service.

## Surfstore Interface
`SurfstoreInterfaces.go` also contains interfaces for the BlockStore and the MetadataStore:

```go
type MetaStoreInterface interface {
	// Retrieves the server's FileInfoMap
	GetFileInfoMap(ctx context.Context, _ *emptypb.Empty) (*FileInfoMap, error)

	// Update a file's fileinfo entry
	UpdateFile(ctx context.Context, fileMetaData *FileMetaData) (*Version, error)

	// Get the the BlockStore address
	GetBlockStoreAddr(ctx context.Context, _ *emptypb.Empty) (*BlockStoreAddr, error)
}

type BlockStoreInterface interface {
	// Get a block based on blockhash
	GetBlock(ctx context.Context, blockHash *BlockHash) (*Block, error)

	// Put a block
	PutBlock(ctx context.Context, block *Block) (*Success, error)

	// Given a list of hashes “in”, returns a list containing the
	// subset of in that are stored in the key-value store
	HasBlocks(ctx context.Context, blockHashesIn *BlockHashes) (*BlockHashes, error)
}
```

## Implementation
### Server
`BlockStore.go` provides a skeleton implementation of the `BlockStoreInterface` and `MetaStore.go` provides a skeleton implementation of the `MetaStoreInterface` 
**You must implement the methods in these 2 files which have `panic("todo")` as their body.**

`cmd/SurfstoreServerExec/main.go` also has a method `startServer` **which you must implement**. Depending on the service type specified, it should register a `MetaStore`, `BlockStore`, or `Both` and start listening for connections from clients.

### Client
`SurfstoreRPCClient.go` provides the gRPC client stub for the surfstore gRPC server. **You must implement the methods in this file which have `panic("todo")` as their body.** (Hint: one of them has been implemented for you)

`SurfstoreUtils.go` also has the following method which **you need to implement** for the sync logic of clients:
```go
/*
Implement the logic for a client syncing with the server here.
*/
func ClientSync(client RPCClient) {
	panic("todo")
}
```
## Usage
1. Run your server using this:
```shell
go run cmd/SurfstoreServerExec/main.go -s <service> -p <port> -l -d (BlockStoreAddr*)
```
Here, `service` should be one of three values: meta, block, or both. This is used to specify the service provided by the server. `port` defines the port number that the server listens to (default=8080). `-l` configures the server to only listen on localhost. `-d` configures the server to output log statements. Lastly, (BlockStoreAddr\*) is the BlockStore address that the server is configured with. If `service=both` then the BlockStoreAddr should be the `ip:port` of this server.

2. Run your client using this:
```shell
go run cmd/SurfstoreClientExec/main.go -d <meta_addr:port> <base_dir> <block_size>
```

## Examples:
```shell
go run cmd/SurfstoreServerExec/main.go -s both -p 8081 -l localhost:8081
```
This starts a server that listens only to localhost on port 8081 and services both the BlockStore and MetaStore interface.

```shell
Run the commands below on separate terminals (or nodes)
> go run cmd/SurfstoreServerExec/main.go -s block -p 8081 -l
> go run cmd/SurfstoreServerExec/main.go -s meta -l localhost:8081
```
The first line starts a server that services only the BlockStore interface and listens only to localhost on port 8081. The second line starts a server that services only the MetaStore interface, listens only to localhost on port 8080, and references the BlockStore we created as the underlying BlockStore. (Note: if these are on separate nodes, then you should use the public ip address and remove `-l`)

3. From a new terminal (or a new node), run the client using the script provided in the starter code (if using a new node, build using step 1 first). Use a base directory with some files in it.
```shell
> mkdir dataA
> cp ~/pic.jpg dataA/ 
> go run cmd/SurfstoreClientExec/main.go server_addr:port dataA 4096
```
This would sync pic.jpg to the server hosted on `server_addr:port`, using `dataA` as the base directory, with a block size of 4096 bytes.

4. From another terminal (or a new node), run the client to sync with the server. (if using a new node, build using step 1 first)
```shell
> mkdir dataB
> go run cmd/SurfstoreClientExec/main.go server_addr:port dataB 4096
> ls dataB/
pic.jpg index.txt
```
We observe that pic.jpg has been synced to this client.

## Makefile
We also provide a make file for you to run the BlockStore and MetaStore servers.
1. Run both BlockStore and MetaStore servers (**listens to localhost on port 8081**):
```shell
make run-both
```

2. Run BlockStore server (**listens to localhost on port 8081**):
```shell
make run-blockstore
```

3. Run MetaStore server (**listens to localhost on port 8080**):
```shell
make run-metastore
```

## Testing 
On gradescope, only a subset of test cases will be visible, so we highly encourage you to come up with different scenarios like the one described above. You can then match the outcome of your implementation to the expected output based on the theory provided in the writeup.
