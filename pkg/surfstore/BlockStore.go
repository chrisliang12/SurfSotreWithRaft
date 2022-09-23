package surfstore

import (
	context "context"
	"fmt"
)

type BlockStore struct {
	BlockMap map[string]*Block
	UnimplementedBlockStoreServer
}

func (bs *BlockStore) GetBlock(ctx context.Context, blockHash *BlockHash) (*Block, error) {
	b, ok := bs.BlockMap[blockHash.Hash]
	if !ok {
		return nil, fmt.Errorf("cannot get block")
	}
	return b, nil
}

func (bs *BlockStore) PutBlock(ctx context.Context, block *Block) (*Success, error) {
	hashVal := GetBlockHashString(block.BlockData)
	bs.BlockMap[hashVal] = block
	return &Success{Flag: true}, nil
}

// Given a list of hashes “in”, returns a list containing the
// subset of in that are stored in the key-value store
func (bs *BlockStore) HasBlocks(ctx context.Context, blockHashesIn *BlockHashes) (*BlockHashes, error) {
	hashin := blockHashesIn.Hashes
	var hashout []string
	for _, hash := range hashin {
		if _, ok := bs.BlockMap[hash]; ok {
			hashout = append(hashout, hash)
		}
	}
	if len(hashout) == 0 {
		return nil, fmt.Errorf("cannot find blocks")
	}
	return &BlockHashes{Hashes: hashout}, nil
}

// This line guarantees all method for BlockStore are implemented
var _ BlockStoreInterface = new(BlockStore)

func NewBlockStore() *BlockStore {
	return &BlockStore{
		BlockMap: map[string]*Block{},
	}
}
