package prosl_test

import (
	"encoding/hex"
	"github.com/proskenion/proskenion/core"
	"github.com/proskenion/proskenion/core/model"
	. "github.com/proskenion/proskenion/prosl"
	"github.com/proskenion/proskenion/query"
	"github.com/proskenion/proskenion/repository"
	. "github.com/proskenion/proskenion/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

const genesisRootId = "root@com"

func Initalize() (core.Repository, model.ModelFactory, core.QueryProcessor, core.QueryValidator, core.QueryVerifier) {
	dba := RandomDBA()
	cryptor := RandomCryptor()
	fc := NewTestFactory()
	rp := repository.NewRepository(dba, cryptor, fc)
	conf := NewTestConfig()

	qp := query.NewQueryProcessor(rp, fc, conf)
	qv := query.NewQueryValidator(rp, fc, conf)
	qc := query.NewQueryVerifier()
	return rp, fc, qp, qv, qc
}

func NewQuerycutor(qp core.QueryProcessor, qv core.QueryValidator, qc core.QueryVerifier) core.Querycutor {
	return struct {
		core.QueryProcessor
		core.QueryValidator
		core.QueryVerifier
	}{qp, qv, qc}
}

var (
	authorizer AccountWithPri
	peer       PeerWithPri
	acs        []AccountWithPri
)

func FileToBianry(t *testing.T, filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	ret, err := hex.DecodeString(string(data))
	require.NoError(t, err)
	return ret
}

type KeyPair struct {
	model.PublicKey
	model.PrivateKey
}

func InitializeObjects(t *testing.T) {
	keypairs := make([]*KeyPair, 0)
	for _, name := range []string{
		"key_auth", "key_peer", "key1", "key2", "key3",
	} {
		pub := FileToBianry(t, "./test_yaml/"+name+".pub")
		pri := FileToBianry(t, "./test_yaml/"+name+".pri")
		keypairs = append(keypairs, &KeyPair{pub, pri})
	}
	authorizer = AccountWithPri{
		"authorizer@com",
		keypairs[0].PublicKey,
		keypairs[0].PrivateKey,
	}
	peer = PeerWithPri{
		NewTestFactory().NewPeer("root@peer", "127.0.0.1:50055", keypairs[1].PublicKey),
		keypairs[1].PrivateKey,
	}
	acs = []AccountWithPri{
		{
			"account1@com",
			keypairs[2].PublicKey,
			keypairs[2].PrivateKey,
		},
		{
			"account2@com",
			keypairs[3].PublicKey,
			keypairs[3].PrivateKey,
		},
		{
			"account3@com",
			keypairs[4].PublicKey,
			keypairs[4].PrivateKey,
		},
	}
}

func testGenesisExecuteProsl(t *testing.T, filename string, value *ProslStateValue, rp core.Repository) {
	buf, err := ioutil.ReadFile(filename)
	require.NoError(t, err)

	prosl, err := ConvertYamlToProtobuf(buf)
	require.NoError(t, err)

	state := ExecuteProsl(prosl, value)
	require.NoError(t, state.Err)
	require.NotNil(t, state.ReturnObject)

	expB := NewTestFactory().NewTxBuilder().
		AddPeer(genesisRootId,
			peer.GetPeerId(),
			peer.GetAddress(),
			peer.GetPublicKey()).
		CreateAccount(genesisRootId, authorizer.AccountId, []model.PublicKey{authorizer.Pubkey}, 1)
	for _, ac := range acs {
		expB = expB.CreateAccount(genesisRootId, ac.AccountId, []model.PublicKey{ac.Pubkey}, 1)
	}
	for i, ac := range acs {
		expB = expB.AddBalance(genesisRootId, ac.AccountId, int64(10000*(i+1)))
	}
	expTx := expB.Build()
	actualTx := state.ReturnObject.GetTransaction()
	assert.Equal(t, expTx.Hash(), actualTx.Hash())

	txList := EmptyTxList()
	require.NoError(t, txList.Push(actualTx))
	rp.GenesisCommit(txList)
}

func testGetAccountsExecuteProsl(t *testing.T, filename string, value *ProslStateValue) {
	buf, err := ioutil.ReadFile(filename)
	require.NoError(t, err)

	prosl, err := ConvertYamlToProtobuf(buf)
	require.NoError(t, err)

	state := ExecuteProsl(prosl, value)
	require.NoError(t, state.Err)
	require.NotNil(t, state.ReturnObject)

	exIds := make([]string, 3)
	for i, ac := range acs {
		exIds[2-i] = ac.AccountId
	}
	actualList := state.ReturnObject.GetList()
	for i, id := range exIds {
		assert.Equal(t, id, actualList[i].GetAccount().GetAccountId())
	}
}

func TestExecuteProsl(t *testing.T) {
	rp, fc, qp, qv, qc := Initalize()
	InitializeObjects(t)
	testGenesisExecuteProsl(t, "./test_yaml/genesis.yaml",
		InitProslStateValue(fc, NewQuerycutor(qp, qv, qc)), rp)

	testGetAccountsExecuteProsl(t, "./test_yaml/test_1.yaml",
		InitProslStateValue(fc, NewQuerycutor(qp, qv, qc)))

}
