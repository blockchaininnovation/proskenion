package core

import (
	"github.com/pkg/errors"
	. "github.com/proskenion/proskenion/core/model"
)

var (
	ErrMerklePatriciaTreeNotFoundKey = errors.Errorf("Failed MerklePatriciaTree Not Found key")
	ErrInvalidKVNodes                = errors.Errorf("Failed Key Value Nodes Invalid")
	ErrWSVNotFound                   = errors.Errorf("Failed WSV Query Not Found")
	ErrWSVQueryUnmarshal             = errors.Errorf("Failed WSV Query Unmarshal")
	ErrTxHistoryNotFound             = errors.Errorf("Failed TxHistory Query Not Found")
	ErrTxHistoryQueryUnmarshal       = errors.Errorf("Failed TxHistory Query Unmarshal")
	ErrBlockchainNotFound            = errors.Errorf("Failed Blockchain Get Not Found")
	ErrBlockchainQueryUnmarshal      = errors.Errorf("Failed Blocchain Get Unmarshal")
)

var (
	ErrRepositoryCommitLoadPreBlock   = errors.Errorf("Failed Repository Commit Load PreBlockchain")
	ErrRepositoryCommitLoadWSV        = errors.Errorf("Failed Repository Commit Load WSV")
	ErrRepositoryCommitLoadTxHistory  = errors.Errorf("Failed Repository Commit Load TxHistory")
)

// Transaction 列の管理
type MerkleTree interface {
	Push(hash Hasher) error
	Top() Hash
}

// TxList Wrap MerkleTree
type TxList interface {
	Push(tx Transaction) error
	Top() Hash
	List() []Transaction
	Size() int
}

type KVNode interface {
	// KVNode{key = Key()[cnt:], value=value}
	Next(cnt int) KVNode
	Key() []byte
	Value() Marshaler
}

// Merkle Patricia Tree に対する操作
type MerklePatriciaController interface {
	// key で参照した先の iterator を取得
	Find(key []byte) (MerklePatriciaNodeIterator, error)
	// Upsert したあとの Iterator を生成して取得
	Upsert(KVNode) (MerklePatriciaNodeIterator, error)
	// hash を Root にする
	Set(hash Hash) error
	Hasher
	Marshaler
	Unmarshaler
}

// World State の管理 に使う(SubTree の管理にも使う)
type MerklePatriciaTree interface {
	Iterator() MerklePatriciaNodeIterator
	MerklePatriciaController
}

// Merkle Patricia Node を管理する Iterator
type MerklePatriciaNodeIterator interface {
	MerklePatriciaController
	Key() []byte
	Childs() map[byte]Hash
	DataHash() Hash
	Leaf() bool
	Data(unmarshaler Unmarshaler) error
	Prev() (MerklePatriciaNodeIterator, error)
}

// WSV (MerklePatriciaTree で管理)
type WSV interface {
	Hash() (Hash, error)
	// Query gets value from targetId
	Query(targetId string, value Unmarshaler) error
	// Append [targetId] = value
	Append(targetId string, value Marshaler) error
	// Commit appenging nodes
	Commit() error
	// RollBack
	Rollback() error
}

// 全Tx履歴 (MerklePatriciaTree で管理)
type TxHistory interface {
	Hash() (Hash, error)
	// Query gets tx from txHash
	Query(txHash Hash) (Transaction, error)
	// Append tx
	Append(tx Transaction) error
	// Commit appenging nodes
	Commit() error
	// RollBack
	Rollback() error
}

// BlockChain
type Blockchain interface {
	// blockHash を指定して Block を取得
	Get(blockHash Hash) (Block, error)
	// Commit block
	Append(block Block) error
}

// 提案された Transaction を保持する Queue
type ProposalTxQueue interface {
	Push(tx Transaction) error
	Erase(hash Hash) error
	Pop() (Transaction, bool)
}

type Repository interface {
	Begin() (RepositoryTx, error)
	Top() (Block, bool)
	Commit(Block, TxList) error
	GeneisCommit(TxList) error
}

type RepositoryTx interface {
	WSV(Hash) (WSV, error)
	TxHistory(Hash) (TxHistory, error)
	Blockchain(Hash) (Blockchain, error)
	Commit() error
	Rollback() error
}
