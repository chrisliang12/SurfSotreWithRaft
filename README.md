# Instructions to extend project 4

1. Make a copy of your solution if you want to:
```console
mkdir proj5
cp -r proj4/* proj5
cd proj5
```

2. Rename all module paths from "proj4" to "proj5" (you may have more that are not shown here)
```console
$ grep -r proj4 ./
cmd/SurfstoreServerExec/main.go:        "cse224/proj4/pkg/surfstore"
cmd/SurfstoreClientExec/main.go:        "cse224/proj4/pkg/surfstore"
go.mod:module cse224/proj4
```

3. Copy over the given test cases
```console
mkdir test
cp -r /path/to/proj5/starter-code/test/* test/
```

4. Copy over the Makefile and example config
```console
cp /path/to/proj5/starter-code/Makefile .
cp /path/to/proj5/starter-code/example_config.txt .
```

5. Copy over Raft specific files
```console
mkdir cmd/SurfstoreRaftServerExec
cp /path/to/proj5/starter-code/cmd/SurfstoreRaftServerExec/main.go cmd/SurfstoreRaftServerExec/
cp /path/to/proj5/starter-code/pkg/surfstore/Raft* pkg/surfstore/
cp /path/to/proj5/starter-code/pkg/surfstore/SurfStore.proto pkg/surfstore/
```

6. Copy over new client exec program and make changes to the client
```console
cp /path/to/proj5/starter-code/cmd/SurfstoreClientExec/main.go cmd/SurfstoreClientExec/
```

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

MetaStore functionality is now provided by the RaftSurfstoreServer, so change the MetaStore clients to RaftSurfstoreServer clients:

```go
c := NewRaftSurfstoreClient(conn)
```

And since we no longer have the `MetaStoreAddr` field, for now you can change `surfclient.MetaStoreAddr` to `surfclient.MetaStoreAddrs[0]`. You will eventually need to change this so you can find a leader, deal with server crashes, etc. 
```go
conn, err := grpc.Dial(surfClient.MetaStoreAddrs[0], grpc.WithInsecure())
```


7. Re-generate the protobuf
```console
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/surfstore/SurfStore.proto
```


You should now be able to run `make test` and it will fail with the panic messages.


# Makefile

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
