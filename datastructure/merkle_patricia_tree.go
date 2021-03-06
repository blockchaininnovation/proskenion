package datastructure

import (
	"bytes"
	"encoding/gob"
	"github.com/pkg/errors"
	"github.com/proskenion/proskenion/core"
	"github.com/proskenion/proskenion/core/model"
	"strconv"
)

// World State の管理
type MerklePatriciaTree struct {
	dba     core.KeyValueStore
	cryptor core.Cryptor
	root    core.MerklePatriciaNodeIterator
}

func makeInitializeMerklePatriciaNodeItreator(kvStore core.KeyValueStore, cryptor core.Cryptor, rootKey byte) *MerklePatriciaNodeIterator {
	return &MerklePatriciaNodeIterator{
		dba:     kvStore,
		cryptor: cryptor,
		node: &MerklePatriciaInternalNode{
			Key_:      []byte{rootKey},
			Childs_:   make(map[byte]model.Hash),
			DataHash_: model.Hash(nil),
		},
	}
}

func NewMerklePatriciaTree(kvStore core.KeyValueStore, cryptor core.Cryptor, hash model.Hash, rootKey byte) (core.MerklePatriciaTree, error) {
	newInternal := &MerklePatriciaNodeIterator{
		dba:     kvStore,
		cryptor: cryptor,
		node:    &MerklePatriciaInternalNode{},
	}
	if hash == nil {
		newInternal = makeInitializeMerklePatriciaNodeItreator(kvStore, cryptor, rootKey)
		hash = newInternal.Hash()
	}
	err := kvStore.Load(hash, newInternal)
	if err != nil {
		if errors.Cause(err) != core.ErrDBANotFoundLoad {
			return nil, err
		}

		// ROOT Internal Node
		newInternal = makeInitializeMerklePatriciaNodeItreator(kvStore, cryptor, rootKey)
		hash := newInternal.Hash()

		// saved
		err = kvStore.Store(hash, newInternal)
		if err != nil {
			return nil, err
		}
	}

	return &MerklePatriciaTree{
		dba:     kvStore,
		cryptor: cryptor,
		root:    newInternal,
	}, nil
}

func (t *MerklePatriciaTree) Iterator() core.MerklePatriciaNodeIterator {
	return t.root
}

// Hash に root を変更
func (t *MerklePatriciaTree) Set(hash model.Hash) error {
	return t.Iterator().Set(hash)
}

func (t *MerklePatriciaTree) Get(hash model.Hash) (core.MerklePatriciaNodeIterator, error) {
	return t.Iterator().Get(hash)
}

// key と prefix が一致している最も浅い internal iterator を取得
func (t *MerklePatriciaTree) Search(key []byte) (core.MerklePatriciaNodeIterator, error) {
	ret, err := t.Iterator().Search(key)
	if err != nil {
		return nil, errors.Wrap(core.ErrMerklePatriciaTreeNotSearchKey, err.Error())
	}
	return ret, nil
}

// key で参照した先の iterator を取得
func (t *MerklePatriciaTree) Find(key []byte) (core.MerklePatriciaNodeIterator, error) {
	ret, err := t.Iterator().Find(key)
	if err != nil {
		return nil, errors.Wrap(core.ErrMerklePatriciaTreeNotFoundKey, err.Error())
	}
	return ret, nil
}

// Upsert したあとの新しい Iterator を生成して取得
func (t *MerklePatriciaTree) Upsert(node core.KVNode) (core.MerklePatriciaNodeIterator, error) {
	it, err := t.Iterator().Upsert(node)
	if err != nil {
		return nil, err
	}
	t.root = it
	return it, nil
}

func (t *MerklePatriciaTree) Hash() model.Hash {
	return t.Iterator().Hash()
}

func (t *MerklePatriciaTree) Marshal() ([]byte, error) {
	return t.Iterator().Marshal()
}

func (t *MerklePatriciaTree) Unmarshal(b []byte) error {
	return t.Iterator().Unmarshal(b)
}

type MerklePatriciaNode interface {
	Leaf() bool

	// Internal
	Key() []byte
	Childs() map[byte]model.Hash
	DataHash() model.Hash

	// Leaf
	Height() int64
	PrevHash() model.Hash
	DataObject() []byte
}

type MerklePatriciaInternalNode struct {
	Key_      []byte
	Childs_   map[byte]model.Hash // child node of tree.
	DataHash_ model.Hash          // data access key (must be leaf node)
}

func (n *MerklePatriciaInternalNode) Leaf() bool {
	return false
}

func (n *MerklePatriciaInternalNode) Key() []byte {
	return n.Key_
}

func (n *MerklePatriciaInternalNode) Childs() map[byte]model.Hash {
	return n.Childs_
}

func (n *MerklePatriciaInternalNode) DataHash() model.Hash {
	return n.DataHash_
}

func (n *MerklePatriciaInternalNode) Height() int64 {
	return 0
}

func (n *MerklePatriciaInternalNode) PrevHash() model.Hash {
	return model.Hash(nil)
}

func (n *MerklePatriciaInternalNode) DataObject() []byte {
	return nil
}

type MerklePatriciaLeafNode struct {
	Height_     int64
	PrevHash_   model.Hash
	DataObject_ []byte // Unmarshaled data object
}

func (n *MerklePatriciaLeafNode) Leaf() bool {
	return true
}

func (n *MerklePatriciaLeafNode) Key() []byte {
	return nil
}

func (n *MerklePatriciaLeafNode) Childs() map[byte]model.Hash {
	return nil
}

func (n *MerklePatriciaLeafNode) DataHash() model.Hash {
	return nil
}

func (n *MerklePatriciaLeafNode) Height() int64 {
	return n.Height_
}

func (n *MerklePatriciaLeafNode) PrevHash() model.Hash {
	return n.PrevHash_
}

func (n *MerklePatriciaLeafNode) DataObject() []byte {
	return n.DataObject_
}

type MerklePatriciaNodeIterator struct {
	dba     core.KeyValueStore
	cryptor core.Cryptor
	node    MerklePatriciaNode
}

// new** は単に型の生成、データの保存は行わない
func (t *MerklePatriciaNodeIterator) newEmptyLeafIterator() core.MerklePatriciaNodeIterator {
	return &MerklePatriciaNodeIterator{
		dba:     t.dba,
		cryptor: t.cryptor,
		node:    &MerklePatriciaLeafNode{},
	}
}

func (t *MerklePatriciaNodeIterator) newEmptyInternalIterator() core.MerklePatriciaNodeIterator {
	return &MerklePatriciaNodeIterator{
		dba:     t.dba,
		cryptor: t.cryptor,
		node:    &MerklePatriciaInternalNode{},
	}
}

func (t *MerklePatriciaNodeIterator) getChild(key byte) (core.MerklePatriciaNodeIterator, error) {
	unmarshaler := t.newEmptyInternalIterator()
	childHash, ok := t.Childs()[key]
	if !ok {
		return nil, errors.Errorf("Childs Out of Range")
	}
	err := t.dba.Load(childHash, unmarshaler)
	if err != nil {
		return nil, err
	}
	return unmarshaler, nil
}

func (t *MerklePatriciaNodeIterator) getLeaf() (core.MerklePatriciaNodeIterator, error) {
	unmarshaler := t.newEmptyLeafIterator()
	err := t.dba.Load(t.DataHash(), unmarshaler)
	if err != nil {
		return nil, err
	}
	return unmarshaler, nil
}

func (t *MerklePatriciaNodeIterator) Data(unmarshaler model.Unmarshaler) error {
	return unmarshaler.Unmarshal(t.node.DataObject())
}

// create ** はデータを保存する
func (t *MerklePatriciaNodeIterator) createMerklePatriciaNodeIterator(node MerklePatriciaNode) (core.MerklePatriciaNodeIterator, error) {
	it := &MerklePatriciaNodeIterator{
		dba:     t.dba,
		cryptor: t.cryptor,
		node:    node,
	}
	hash := it.Hash()
	// saved
	err := t.dba.Store(hash, it)
	if err != nil {
		if errors.Cause(err) == core.ErrDBADuplicateStore { // Duplicate, but return this node
			return it, nil
		}
		return nil, err
	}
	return it, nil
}

// 追加したい node 情報から新しい葉ノード(or node にまだ key が残っているなら interanl node)のイテレータを保存して返す
func (t *MerklePatriciaNodeIterator) createLeafIterator(node core.KVNode) (core.MerklePatriciaNodeIterator, error) {
	object_, err := node.Value().Marshal()
	if err != nil {
		return nil, err
	}
	newLeafIt, err := t.createMerklePatriciaNodeIterator(&MerklePatriciaLeafNode{
		Height_:     0,
		DataObject_: object_,
		PrevHash_:   model.Hash(nil),
	})
	if err != nil {
		return nil, err
	}

	if len(node.Key()) > 0 {
		hash := newLeafIt.Hash()
		newInternalIt, err := t.createMerklePatriciaNodeIterator(
			&MerklePatriciaInternalNode{
				Key_:      node.Key(),
				DataHash_: hash,
				Childs_:   make(map[byte]model.Hash),
			},
		)
		if err != nil {
			return nil, err
		}
		return newInternalIt, nil
	}
	return newLeafIt, nil
}

// key と prefix が一致している最も浅い internal iterator を取得
func (t *MerklePatriciaNodeIterator) Search(key []byte) (core.MerklePatriciaNodeIterator, error) {
	if t.Leaf() {
		return t, nil
	}
	if len(t.Key()) >= len(key) {
		if !bytes.HasPrefix(t.Key(), key) {
			return nil, core.ErrMerklePatriciaTreeNotSearchKey
		}
		return t, nil
	}
	nextChild, err := t.getChild(key[len(t.Key())])
	if err != nil {
		return nil, err
	}
	return nextChild.Search(key[len(t.Key()):])
}

// key で参照した先の iterator を取得
func (t *MerklePatriciaNodeIterator) Find(key []byte) (core.MerklePatriciaNodeIterator, error) {
	if t.Leaf() {
		return t, nil
	}
	if len(t.Key()) > len(key) {
		return nil, core.ErrMerklePatriciaTreeNotFoundKey
	}
	if len(t.Key()) == len(key) {
		if !bytes.Equal(t.Key(), key) {
			return nil, core.ErrMerklePatriciaTreeNotFoundKey
		}
		leaf, err := t.getLeaf()
		if err != nil {
			return nil, err
		}
		return leaf, nil
	}
	nextChild, err := t.getChild(key[len(t.Key())])
	if err != nil {
		return nil, err
	}
	return nextChild.Find(key[len(t.Key()):])
}

// return ( number of match prefix bytes, Is perfect matche? )
func CountPrefixBytes(a []byte, b []byte) (int, bool) {
	cnt := 0
	for ; cnt < len(a) && cnt < len(b); cnt++ {
		if a[cnt] != b[cnt] {
			return cnt, false
		}
	}
	if len(a) == len(b) {
		return cnt, true
	}
	return cnt, false
}

// node の Key に沿って子を返す
func (t *MerklePatriciaNodeIterator) getChildFromNode(node core.KVNode) (core.MerklePatriciaNodeIterator, bool) {
	if len(node.Key()) == 0 {
		return nil, false
	}
	it, err := t.getChild(node.Key()[0])
	if err != nil {
		return nil, false
	}
	return it, true
}

// ノード t を cnt 番目で分割、新たに child を加えた時の 中間ノードの生成(場合により child にも変更を加える)
func (t *MerklePatriciaNodeIterator) createInternalIterator(cnt int, child core.MerklePatriciaNodeIterator) (core.MerklePatriciaNodeIterator, error) {
	// 子ノードを更新
	var err error
	var newDataHash model.Hash = nil
	newChilds := make(map[byte]model.Hash)
	newKey := t.Key()[:cnt]
	childs := []core.MerklePatriciaNodeIterator{child}
	if len(t.Key()) == cnt { // 自身が分裂していない場合は自分の情報を受け継ぐ
		newDataHash = t.DataHash()
		newChilds = t.Childs()
		if child.Leaf() {
			newDataHash = child.Hash()
		}
	} else { // 分裂するときは分裂後の子を生成する
		// 分岐後の自分(child)
		it, err := t.createMerklePatriciaNodeIterator(
			&MerklePatriciaInternalNode{
				Key_:      t.Key()[cnt:],
				Childs_:   t.Childs(),
				DataHash_: t.DataHash(),
			},
		)
		if err != nil {
			return nil, err
		}

		if child.Leaf() {
			// 追加する子が葉であるとき、葉は DataHash に、自分のみを子にする
			newDataHash = child.Hash()
			childs = []core.MerklePatriciaNodeIterator{it}
		} else {
			// そうでなければ自分の分身を新しい子の集合に加える
			childs = append(childs, it)
		}
	}

	for _, child := range childs {
		if child.Leaf() {
			continue
		}
		hash := child.Hash()
		if err != nil {
			return nil, err
		}
		newChilds[child.Key()[0]] = hash
	}

	return t.createMerklePatriciaNodeIterator(
		&MerklePatriciaInternalNode{
			Key_:      newKey,
			Childs_:   newChilds,
			DataHash_: newDataHash,
		})
}

// Upsert したあとの Iterator を生成して取得
func (t *MerklePatriciaNodeIterator) Upsert(node core.KVNode) (core.MerklePatriciaNodeIterator, error) {
	if t.Leaf() {
		return t.Append(node.Value())
	}
	cnt, ok := CountPrefixBytes(t.Key(), node.Key())
	node = node.Next(cnt)

	// key と prefix が完全一致して且つ子ノードが存在する
	if ok { // Perfect Match
		// key と完全一致したので dataHash の中身を更新
		it, err := t.getLeaf()
		if err != nil {
			return nil, err
		}
		newIt, err := it.Upsert(node)
		if err != nil {
			return nil, err
		}
		return t.createInternalIterator(cnt, newIt)
	}
	if len(t.Key()) == cnt {
		if it, ok := t.getChildFromNode(node); ok {
			newIt, err := it.Upsert(node)
			if err != nil {
				return nil, err
			}
			return t.createInternalIterator(cnt, newIt)
		}
	}
	// 現在のノードを分割して中身を更新
	newIt, err := t.createLeafIterator(node)
	if err != nil {
		return nil, err
	}
	return t.createInternalIterator(cnt, newIt)

}

// 現在参照しているノードに値を追加
func (t *MerklePatriciaNodeIterator) Append(value model.Marshaler) (core.MerklePatriciaNodeIterator, error) {
	object, err := value.Marshal()
	if err != nil {
		return nil, err
	}
	thisHash := t.Hash()
	if err != nil {
		return nil, err
	}
	newIt, err := t.createMerklePatriciaNodeIterator(
		&MerklePatriciaLeafNode{
			Height_:     t.node.Height() + 1,
			DataObject_: object,
			PrevHash_:   thisHash,
		},
	)
	if err != nil {
		return nil, err
	}
	return newIt, nil
}

func (t *MerklePatriciaNodeIterator) Leaf() bool {
	return t.node.Leaf()
}

func (t *MerklePatriciaNodeIterator) Prev() (core.MerklePatriciaNodeIterator, error) {
	it := t.newEmptyLeafIterator()
	err := t.dba.Load(t.node.PrevHash(), it)
	if err != nil {
		return nil, err
	}
	return it, nil
}

func (t *MerklePatriciaNodeIterator) Hash() model.Hash {
	if t.Leaf() {
		return t.cryptor.ConcatHash(t.node.PrevHash(),
			t.node.DataObject(),
			[]byte(strconv.FormatInt(t.node.Height(), 10)))
	} else {
		childs := make([]model.Hash, 256)
		for k, v := range t.node.Childs() {
			childs[k] = v
		}
		hash := t.cryptor.ConcatHash(childs...)
		return t.cryptor.ConcatHash(hash, t.node.DataHash(), t.Key())
	}
}

func (t *MerklePatriciaNodeIterator) Marshal() ([]byte, error) {
	var network bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&network) // Will write to network.
	err := enc.Encode(t.node)
	if err != nil {
		return nil, err
	}
	return network.Bytes(), nil
}

func (t *MerklePatriciaNodeIterator) Unmarshal(b []byte) error {
	network := bytes.NewBuffer(b)
	dec := gob.NewDecoder(network) // Will read from network.
	err := dec.Decode(t.node)
	if err != nil {
		return err
	}
	return nil
}

func (t *MerklePatriciaNodeIterator) Key() []byte {
	return t.node.Key()
}

func (t *MerklePatriciaNodeIterator) Childs() map[byte]model.Hash {
	return t.node.Childs()
}

func (t *MerklePatriciaNodeIterator) DataHash() model.Hash {
	return t.node.DataHash()
}

// Hash に現在の Iterator を変更
func (t *MerklePatriciaNodeIterator) Set(hash model.Hash) error {
	return t.dba.Load(hash, t)
}

func (t *MerklePatriciaNodeIterator) Get(hash model.Hash) (core.MerklePatriciaNodeIterator, error) {
	it := t.newEmptyInternalIterator()
	err := t.dba.Load(hash, it)
	if err != nil {
		return nil, err
	}
	return it, nil
}

func (t *MerklePatriciaNodeIterator) SubLeafs() ([]core.MerklePatriciaNodeIterator, error) {
	if t.node.Leaf() {
		return []core.MerklePatriciaNodeIterator{t}, nil
	} else {
		rets := make([]core.MerklePatriciaNodeIterator, 0, len(t.Childs()))
		if len(t.DataHash()) != 0 {
			it, err := t.getLeaf()
			if err != nil {
				return nil, err
			}
			its, err := it.SubLeafs()
			rets = append(rets, its...)
		}
		for key, _ := range t.Childs() {
			it, err := t.getChild(key)
			if err != nil {
				return nil, err
			}
			its, err := it.SubLeafs()
			if err != nil {
				return nil, err
			}
			rets = append(rets, its...)
		}
		return rets, nil
	}
}
