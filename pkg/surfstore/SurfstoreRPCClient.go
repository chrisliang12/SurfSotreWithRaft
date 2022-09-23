package surfstore

import (
	context "context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	grpc "google.golang.org/grpc"
)

type RPCClient struct {
	MetaStoreAddr []string
	BaseDir       string
	BlockSize     int
}

func (surfClient *RPCClient) GetBlock(blockHash string, blockStoreAddr string, block *Block) error {
	// connect to the server
	conn, err := grpc.Dial(blockStoreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c := NewBlockStoreClient(conn)

	// perform the call
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	b, err := c.GetBlock(ctx, &BlockHash{Hash: blockHash})
	if err != nil {
		conn.Close()
		return err
	}
	block.BlockData = b.BlockData
	block.BlockSize = b.BlockSize

	// close the connection
	return conn.Close()
}

func (surfClient *RPCClient) PutBlock(block *Block, blockStoreAddr string, succ *bool) error {
	// connect to the server
	conn, err := grpc.Dial(blockStoreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c := NewBlockStoreClient(conn)

	// perform the call
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ifSucc, err := c.PutBlock(ctx, block)
	*succ = ifSucc.Flag
	if err != nil {
		conn.Close()
		return err
	}
	return conn.Close()
}

func (surfClient *RPCClient) HasBlocks(blockHashesIn []string, blockStoreAddr string, blockHashesOut *[]string) error {
	conn, err := grpc.Dial(blockStoreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c := NewBlockStoreClient(conn)

	// perform the call
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	newBlockHashes := &BlockHashes{Hashes: blockHashesIn}
	returnBlockHashes, err := c.HasBlocks(ctx, newBlockHashes)
	if err != nil {
		conn.Close()
		return err
	}
	*blockHashesOut = returnBlockHashes.Hashes
	return conn.Close()

}

func (surfClient *RPCClient) GetFileInfoMap(serverFileInfoMap *map[string]*FileMetaData) error {
	for _, metaStore := range surfClient.MetaStoreAddr {
		conn, err := grpc.Dial(metaStore, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		c := NewRaftSurfstoreClient(conn)

		// perform the call
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		returnFileInfoMap, err := c.GetFileInfoMap(ctx, &emptypb.Empty{})

		if err != nil {
			if strings.Contains(err.Error(), ERR_SERVER_CRASHED.Error()) || strings.Contains(err.Error(), ERR_NOT_LEADER.Error()) {
				continue
			}
			conn.Close()
			return err
		}
		*serverFileInfoMap = returnFileInfoMap.FileInfoMap
		return conn.Close()
	}
	return errors.New("ERR (Cluster down): Cannot Get FileInfoMap")
}

func (surfClient *RPCClient) UpdateFile(fileMetaData *FileMetaData, latestVersion *int32) error {
	for _, metaStore := range surfClient.MetaStoreAddr {
		// make connection
		conn, err := grpc.Dial(metaStore, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		c := NewRaftSurfstoreClient(conn)

		// perform the call
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		returnVersion, err := c.UpdateFile(ctx, fileMetaData)

		if err != nil {
			if strings.Contains(err.Error(), ERR_SERVER_CRASHED.Error()) || strings.Contains(err.Error(), ERR_NOT_LEADER.Error()) {
				continue
			}
			conn.Close()
			return err
		}
		*latestVersion = returnVersion.Version
		return conn.Close()
	}
	return errors.New("ERR (Cluster down): Cannot Update MetaStore")
}

func (surfClient *RPCClient) GetBlockStoreAddr(blockStoreAddr *string) error {
	for _, metaStore := range surfClient.MetaStoreAddr {
		// make connection
		conn, err := grpc.Dial(metaStore, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		c := NewRaftSurfstoreClient(conn)

		// perform the call
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		returnBlockStoreAddr, err := c.GetBlockStoreAddr(ctx, &emptypb.Empty{})

		if err != nil {
			if strings.Contains(err.Error(), ERR_SERVER_CRASHED.Error()) || strings.Contains(err.Error(), ERR_NOT_LEADER.Error()) { //TODO: check 2th statement
				continue
			}
			conn.Close()
			return err
		}
		*blockStoreAddr = returnBlockStoreAddr.Addr
		return conn.Close()
	}
	return errors.New("ERR (Cluster down): Cannot Get BlockStore Addr")
}

// This line guarantees all method for RPCClient are implemented
var _ ClientInterface = new(RPCClient)

// Create an Surfstore RPC client
func NewSurfstoreRPCClient(addrs []string, baseDir string, blockSize int) RPCClient {

	return RPCClient{
		MetaStoreAddr: addrs,
		BaseDir:       baseDir,
		BlockSize:     blockSize,
	}
}
