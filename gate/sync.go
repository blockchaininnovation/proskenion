package gate

import (
	"fmt"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/proskenion/proskenion/config"
	"github.com/proskenion/proskenion/core"
	"github.com/proskenion/proskenion/core/model"
)

type SyncGate struct {
	rp     core.Repository
	fc     model.ModelFactory
	c      core.Cryptor
	logger log15.Logger
	conf   *config.Config
}

func NewSyncGate(rp core.Repository, fc model.ModelFactory, c core.Cryptor, logger log15.Logger, conf *config.Config) core.SyncGate {
	return &SyncGate{rp, fc, c, logger, conf}
}

func (c *SyncGate) Sync(blockHash model.Hash, blockChan chan model.Block, txListChan chan core.TxList) error {
	top, ok := c.rp.Top()
	if !ok {
		return fmt.Errorf("top block is empty")
	}
	rtx, err := c.rp.Begin()
	if err != nil {
		return err
	}
	defer core.CommitTx(rtx)
	bc, err := rtx.Blockchain(top.Hash())
	if err != nil {
		return err
	}
	txHistory, err := rtx.TxHistory(top.GetPayload().GetTxHistoryHash())
	if err != nil {
		return err
	}
	for i := 0; i < c.conf.Sync.Limits; i++ {
		block, err := bc.Next(blockHash)
		if err != nil {
			// next がないので正常終了
			if errors.Cause(err) == core.ErrBlockchainNextNotFound {
				return nil
			}
			return err
		}
		blockChan <- block
		txList, err := txHistory.GetTxList(block.GetPayload().GetTxsHash())
		if err != nil {
			return err
		}
		txListChan <- txList
		blockHash = block.Hash()
	}
	return nil
}
