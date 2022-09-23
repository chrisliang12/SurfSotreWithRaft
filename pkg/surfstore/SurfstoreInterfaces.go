package surfstore

import (
	context "context"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

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

type ClientInterface interface {
	// MetaStore
	GetFileInfoMap(serverFileInfoMap *map[string]*FileMetaData) error
	UpdateFile(fileMetaData *FileMetaData, latestVersion *int32) error
	GetBlockStoreAddr(blockStoreAddr *string) error

	// BlockStore
	GetBlock(blockHash string, blockStoreAddr string, block *Block) error
	PutBlock(block *Block, blockStoreAddr string, succ *bool) error
	HasBlocks(blockHashesIn []string, blockStoreAddr string, blockHashesOut *[]string) error
}
