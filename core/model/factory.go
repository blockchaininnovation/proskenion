package model

import (
	"github.com/pkg/errors"
)

var (
	ErrNewBlock    = errors.Errorf("Failed Factory NewBlock")
	ErrNewProposal = errors.Errorf("Failed Factory NewProposal")
)

type ObjectFactory interface {
	NewSignature(pubkey PublicKey, signature []byte) Signature
	NewAccount(accountId string, accountName string, publicKeys []PublicKey, quorum int32, amount int64, peerId string) Account
	NewPeer(peerId string, address string, pubkey PublicKey) Peer

	NewStorageBuilder() StorageBuilder
	NewAccountBuilder() AccountBuilder
	NewObjectBuilder() ObjectBuilder

	NewEmptySignature() Signature
	NewEmptyAccount() Account
	NewEmptyPeer() Peer
	NewEmptyStorage() Storage
	NewEmptyObject() Object
}

type ModelFactory interface {
	ObjectFactory

	NewBlockBuilder() BlockBuilder
	NewTxBuilder() TxBuilder
	NewQueryBuilder() QueryBuilder
	NewQueryResponseBuilder() QueryResponseBuilder

	NewEmptyBlock() Block
	NewEmptyTx() Transaction
	NewEmptyQuery() Query
	NewEmptyQueryResponse() QueryResponse
}

type ObjectBuilder interface {
	Bool(value bool) Object
	Int32(value int32) Object
	Int64(value int64) Object
	Uint32(value uint32) Object
	Uint64(value uint64) Object
	Str(value string) Object
	Data(value []byte) Object
	Address(value string) Object
	Sig(value Signature) Object
	Account(value Account) Object
	Peer(value Peer) Object
	List(value []Object) Object
	Dict(value map[string]Object) Object
	Storage(value Storage) Object
	Command(value Command) Object
	Transaction(value Transaction) Object
	Block(block Block) Object
	Build() Object
}

type StorageBuilder interface {
	From(Storage) StorageBuilder
	FromMap(map[string]Object) StorageBuilder
	Int32(key string, value int32) StorageBuilder
	Int64(key string, value int64) StorageBuilder
	Uint32(key string, value uint32) StorageBuilder
	Uint64(key string, value uint64) StorageBuilder
	Str(key string, value string) StorageBuilder
	Data(key string, value []byte) StorageBuilder
	Address(key string, value string) StorageBuilder
	Sig(key string, value Signature) StorageBuilder
	Account(key string, value Account) StorageBuilder
	Peer(key string, value Peer) StorageBuilder
	List(key string, value []Object) StorageBuilder
	Dict(key string, value map[string]Object) StorageBuilder
	Set(key string, value Object) StorageBuilder
	Id(id string) StorageBuilder
	Build() Storage
}

type AccountBuilder interface {
	From(Account) AccountBuilder
	AccountId(string) AccountBuilder
	AccountName(string) AccountBuilder
	Balance(int64) AccountBuilder
	PublicKeys([]PublicKey) AccountBuilder
	Quorum(int32) AccountBuilder
	DelegatePeerId(string) AccountBuilder
	Build() Account
}

type BlockBuilder interface {
	Height(int64) BlockBuilder
	PreBlockHash(Hash) BlockBuilder
	CreatedTime(int64) BlockBuilder
	WSVHash(Hash) BlockBuilder
	TxHistoryHash(Hash) BlockBuilder
	TxListHash(Hash) BlockBuilder
	Round(int32) BlockBuilder
	Build() Block
}

type TxBuilder interface {
	CreatedTime(int64) TxBuilder
	TransferBalance(authorizerId string, srcAccountId string, destAccountId string, amount int64) TxBuilder
	CreateAccount(authorizerId string, accountId string, publicKeys []PublicKey, quorum int32) TxBuilder
	AddBalance(authorizerId string, accountId string, amount int64) TxBuilder
	AddPublicKeys(authorizerId string, accountId string, pubkeys []PublicKey) TxBuilder
	RemovePublicKeys(authorizerId string, accountId string, pubkeys []PublicKey) TxBuilder
	SetQuorum(authorizerId string, accountId string, quorum int32) TxBuilder
	DefineStorage(authorizerId string, storageId string, storage Storage) TxBuilder
	CreateStorage(authorizerId string, walletId string) TxBuilder
	UpdateObject(authorizerId string, walletId string, key string, object Object) TxBuilder
	AddObject(authorizerId string, walletId string, key string, object Object) TxBuilder
	TransferObject(authorizerId string, walletId string, destAccountId string, key string, object Object) TxBuilder
	AddPeer(authorizerId string, peerId string, address string, pubkey PublicKey) TxBuilder
	ActivatePeer(authorizerId string, peerId string) TxBuilder
	SuspendPeer(authorizerId string, peerId string) TxBuilder
	BanPeer(authorizerId string, peerId string) TxBuilder
	Consign(authorizerId string, accountId string, peerId string) TxBuilder
	CheckAndCommitProsl(authorizerId string, proslId string, params map[string]Object) TxBuilder
	ForceUpdateStorage(authorizerId string, targetId string, storage Storage) TxBuilder
	AppendCommand(cmd Command) TxBuilder
	Build() Transaction
}

type QueryBuilder interface {
	AuthorizerId(string) QueryBuilder
	Select(string) QueryBuilder
	FromId(string) QueryBuilder
	Where(string) QueryBuilder
	OrderBy(key string, order OrderCode) QueryBuilder
	Limit(int32) QueryBuilder
	CreatedTime(int64) QueryBuilder
	RequestCode(code ObjectCode) QueryBuilder
	Build() Query
}

type QueryResponseBuilder interface {
	Account(Account) QueryResponseBuilder
	Peer(Peer) QueryResponseBuilder
	Storage(Storage) QueryResponseBuilder
	List([]Object) QueryResponseBuilder
	Object(Object) QueryResponseBuilder
	Build() QueryResponse
}
