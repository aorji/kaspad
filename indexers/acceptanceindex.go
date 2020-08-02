package indexers

import (
	"bytes"
	"encoding/gob"
	"github.com/kaspanet/kaspad/consensus/accpetancedata"

	"github.com/kaspanet/kaspad/consensus/blockdag"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// AcceptanceIndex implements a txAcceptanceData by block hash index. That is to say,
// it stores a mapping between a block's hash and the set of transactions that the
// block accepts among its blue blocks.
type AcceptanceIndex struct {
	dag             *blockdag.BlockDAG
	databaseContext *dbaccess.DatabaseContext
}

// Ensure the AcceptanceIndex type implements the Indexer interface.
var _ Indexer = (*AcceptanceIndex)(nil)

// NewAcceptanceIndex returns a new instance of an indexer that is used to create a
// mapping between block hashes and their txAcceptanceData.
//
// It implements the Indexer interface which plugs into the IndexManager that in
// turn is used by the blockdag package. This allows the index to be
// seamlessly maintained along with the DAG.
func NewAcceptanceIndex() *AcceptanceIndex {
	return &AcceptanceIndex{}
}

// DropAcceptanceIndex drops the acceptance index.
func DropAcceptanceIndex(databaseContext *dbaccess.DatabaseContext) error {
	dbTx, err := databaseContext.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()

	err = dbaccess.DropAcceptanceIndex(dbTx)
	if err != nil {
		return err
	}

	return dbTx.Commit()
}

// Init initializes the hash-based acceptance index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Init(dag *blockdag.BlockDAG, databaseContext *dbaccess.DatabaseContext) error {
	idx.dag = dag
	idx.databaseContext = databaseContext
	return idx.recover()
}

// recover attempts to insert any data that's missing from the
// acceptance index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) recover() error {
	return idx.dag.ForEachHash(func(hash daghash.Hash) error {
		dbTx, err := idx.databaseContext.NewTx()
		if err != nil {
			return err
		}
		defer dbTx.RollbackUnlessClosed()

		exists, err := dbaccess.HasAcceptanceData(dbTx, &hash)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		txAcceptanceData, err := idx.dag.TxsAcceptedByBlockHash(&hash)
		if err != nil {
			return err
		}
		err = idx.ConnectBlock(dbTx, &hash, txAcceptanceData)
		if err != nil {
			return err
		}

		return dbTx.Commit()
	})
}

// ConnectBlock is invoked by the index manager when a new block has been
// connected to the DAG.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) ConnectBlock(dbContext *dbaccess.TxContext, blockHash *daghash.Hash,
	txsAcceptanceData accpetancedata.MultiBlockTxsAcceptanceData) error {
	serializedTxsAcceptanceData, err := serializeMultiBlockTxsAcceptanceData(txsAcceptanceData)
	if err != nil {
		return err
	}
	return dbaccess.StoreAcceptanceData(dbContext, blockHash, serializedTxsAcceptanceData)
}

// TxsAcceptanceData returns the acceptance data of all the transactions that
// were accepted by the block with hash blockHash.
func (idx *AcceptanceIndex) TxsAcceptanceData(blockHash *daghash.Hash) (accpetancedata.MultiBlockTxsAcceptanceData, error) {
	serializedTxsAcceptanceData, err := dbaccess.FetchAcceptanceData(idx.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}
	return deserializeMultiBlockTxsAcceptanceData(serializedTxsAcceptanceData)
}

type serializableTxAcceptanceData struct {
	MsgTx      wire.MsgTx
	IsAccepted bool
}

type serializableBlockTxsAcceptanceData struct {
	BlockHash        daghash.Hash
	TxAcceptanceData []serializableTxAcceptanceData
}

type serializableMultiBlockTxsAcceptanceData []serializableBlockTxsAcceptanceData

func serializeMultiBlockTxsAcceptanceData(
	multiBlockTxsAcceptanceData accpetancedata.MultiBlockTxsAcceptanceData) ([]byte, error) {
	// Convert MultiBlockTxsAcceptanceData to a serializable format
	serializableData := make(serializableMultiBlockTxsAcceptanceData, len(multiBlockTxsAcceptanceData))
	for i, blockTxsAcceptanceData := range multiBlockTxsAcceptanceData {
		serializableBlockData := serializableBlockTxsAcceptanceData{
			BlockHash:        blockTxsAcceptanceData.BlockHash,
			TxAcceptanceData: make([]serializableTxAcceptanceData, len(blockTxsAcceptanceData.TxAcceptanceData)),
		}
		for i, txAcceptanceData := range blockTxsAcceptanceData.TxAcceptanceData {
			serializableBlockData.TxAcceptanceData[i] = serializableTxAcceptanceData{
				MsgTx:      *txAcceptanceData.Tx.MsgTx(),
				IsAccepted: txAcceptanceData.IsAccepted,
			}
		}
		serializableData[i] = serializableBlockData
	}

	// Serialize
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(serializableData)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func deserializeMultiBlockTxsAcceptanceData(
	serializedTxsAcceptanceData []byte) (accpetancedata.MultiBlockTxsAcceptanceData, error) {
	// Deserialize
	buffer := bytes.NewBuffer(serializedTxsAcceptanceData)
	decoder := gob.NewDecoder(buffer)
	var serializedData serializableMultiBlockTxsAcceptanceData
	err := decoder.Decode(&serializedData)
	if err != nil {
		return nil, err
	}

	// Convert serializable format to MultiBlockTxsAcceptanceData
	multiBlockTxsAcceptanceData := make(accpetancedata.MultiBlockTxsAcceptanceData, len(serializedData))
	for i, serializableBlockData := range serializedData {
		blockTxsAcceptanceData := accpetancedata.BlockTxsAcceptanceData{
			BlockHash:        serializableBlockData.BlockHash,
			TxAcceptanceData: make([]accpetancedata.TxAcceptanceData, len(serializableBlockData.TxAcceptanceData)),
		}
		for i, txData := range serializableBlockData.TxAcceptanceData {
			msgTx := txData.MsgTx
			blockTxsAcceptanceData.TxAcceptanceData[i] = accpetancedata.TxAcceptanceData{
				Tx:         util.NewTx(&msgTx),
				IsAccepted: txData.IsAccepted,
			}
		}
		multiBlockTxsAcceptanceData[i] = blockTxsAcceptanceData
	}

	return multiBlockTxsAcceptanceData, nil
}
