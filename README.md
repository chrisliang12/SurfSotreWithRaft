# Surfstore: A Fault-tolerant Cloud File Storage Service 
This project is the project 5 for CSE 224 Graduate Networked System @ UC San Diego. In this project, we first developed a Dropbox-like cloud-based file storage system with Go and gRPC. Then, we implemented RAFT consensus algorithm with heart beating and log replication to bring fault-tolerance feature to the service. 

# Surfstore without RAFT
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


### Client
`SurfstoreRPCClient.go` provides the gRPC client stub for the surfstore gRPC server.

`SurfstoreUtils.go` also has the following method which **you need to implement** for the sync logic of clients:
```go
/*
Implement the logic for a client syncing with the server here.
*/
func ClientSync(client RPCClient) {
	panic("todo")
}
```




# Surfstore with RAFT
## RAFT Consensus Algorithm
If you are not familiar with RAFT, please see the [RAFT website](https://raft.github.io/)
## MetaStore
With the implementation of RAFT, MetaStore functionality is now provided by the RaftSurfstoreServer.

## Client
The client will need to take a slice of strings instead of a single address. In `pkg/surfstore/SurfstoreRPCClient.go` change the client struct to:

```go
type RPCClient struct {
        MetaStoreAddrs []string
        BaseDir       string
        BlockSize     int
}
```

And change the `NewSurfstoreRPCClient` function to:

```go
func NewSurfstoreRPCClient(addrs []string, baseDir string, blockSize int) RPCClient {
        return RPCClient{
                MetaStoreAddrs: addrs,
                BaseDir:       baseDir,
                BlockSize:     blockSize,
        }
}
```

## Re-generate protobuf
```console
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/surfstore/SurfStore.proto
```

## Makefile

Run BlockStore server:
```console
$ make run-blockstore
```

Run RaftSurfstore server:
```console
$ make IDX=0 run-raft
```

Test:
```console
$ make test
```

Specific Test:
```console
$ make TEST_REGEX=Test specific-test
```

Clean:
```console
$ make clean
```
