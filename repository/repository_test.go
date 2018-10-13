package repository_test

import (
	. "github.com/proskenion/proskenion/repository"
	. "github.com/proskenion/proskenion/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRepository_Commit(t *testing.T) {
	dba := RandomDBA()
	cryptor := RandomCryptor()
	fc := NewTestFactory()

	rp := NewRepository(dba, cryptor, fc)

	block, txList := RandomCommitableBlock(t, nil, rp)
	assert.NoError(t, rp.Commit(block, txList))

	topBlock, ok := rp.Top()
	require.True(t, ok)
	assert.Equal(t, block, topBlock)

	rtx, err := rp.Begin()
	require.NoError(t, err)
	bc, err := rtx.Blockchain(MustHash(topBlock))
	require.NoError(t, err)

	topBlock2, err := bc.Get(MustHash(topBlock))
	require.NoError(t, err)
	assert.Equal(t, MustHash(block), MustHash(topBlock2))
}