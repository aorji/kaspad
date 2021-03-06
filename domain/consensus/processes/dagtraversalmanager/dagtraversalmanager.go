package dagtraversalmanager

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// dagTraversalManager exposes methods for travering blocks
// in the DAG
type dagTraversalManager struct {
	databaseContext model.DBReader

	dagTopologyManager    model.DAGTopologyManager
	ghostdagDataStore     model.GHOSTDAGDataStore
	reachabilityDataStore model.ReachabilityDataStore
	ghostdagManager       model.GHOSTDAGManager
}

// selectedParentIterator implements the `model.BlockIterator` API
type selectedParentIterator struct {
	databaseContext   model.DBReader
	ghostdagDataStore model.GHOSTDAGDataStore
	current           *externalapi.DomainHash
}

func (spi *selectedParentIterator) Next() bool {
	if spi.current == nil {
		return false
	}
	ghostdagData, err := spi.ghostdagDataStore.Get(spi.databaseContext, spi.current)
	if err != nil {
		panic(fmt.Sprintf("ghostdagDataStore is missing ghostdagData for: %v. '%s' ", spi.current, err))
	}
	spi.current = ghostdagData.SelectedParent()
	return spi.current != nil
}

func (spi *selectedParentIterator) Get() *externalapi.DomainHash {
	return spi.current
}

// New instantiates a new DAGTraversalManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	reachabilityDataStore model.ReachabilityDataStore,
	ghostdagManager model.GHOSTDAGManager) model.DAGTraversalManager {
	return &dagTraversalManager{
		databaseContext:       databaseContext,
		dagTopologyManager:    dagTopologyManager,
		ghostdagDataStore:     ghostdagDataStore,
		reachabilityDataStore: reachabilityDataStore,
		ghostdagManager:       ghostdagManager,
	}
}

// SelectedParentIterator creates an iterator over the selected
// parent chain of the given highHash
func (dtm *dagTraversalManager) SelectedParentIterator(highHash *externalapi.DomainHash) model.BlockIterator {
	return &selectedParentIterator{
		databaseContext:   dtm.databaseContext,
		ghostdagDataStore: dtm.ghostdagDataStore,
		current:           highHash,
	}
}

// BlockAtDepth returns the hash of the highest block with a blue score
// lower than (highHash.blueSore - depth) in the selected-parent-chain
// of the block with the given highHash's selected parent chain.
func (dtm *dagTraversalManager) BlockAtDepth(highHash *externalapi.DomainHash, depth uint64) (*externalapi.DomainHash, error) {
	currentBlockHash := highHash
	highBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}

	requiredBlueScore := uint64(0)
	if highBlockGHOSTDAGData.BlueScore() > depth {
		requiredBlueScore = highBlockGHOSTDAGData.BlueScore() - depth
	}

	currentBlockGHOSTDAGData := highBlockGHOSTDAGData
	// If we used `BlockIterator` we'd need to do more calls to `ghostdagDataStore` so we can get the blueScore
	for currentBlockGHOSTDAGData.BlueScore() >= requiredBlueScore {
		if currentBlockGHOSTDAGData.SelectedParent() == nil { // genesis
			return currentBlockHash, nil
		}
		currentBlockHash = currentBlockGHOSTDAGData.SelectedParent()
		currentBlockGHOSTDAGData, err = dtm.ghostdagDataStore.Get(dtm.databaseContext, currentBlockHash)
		if err != nil {
			return nil, err
		}
	}
	return currentBlockHash, nil
}

func (dtm *dagTraversalManager) LowestChainBlockAboveOrEqualToBlueScore(highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error) {
	highBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}

	if highBlockGHOSTDAGData.BlueScore() < blueScore {
		return nil, errors.Errorf("the given blue score %d is higher than block %s blue score of %d",
			blueScore, highHash, highBlockGHOSTDAGData.BlueScore())
	}

	currentHash := highHash
	currentBlockGHOSTDAGData := highBlockGHOSTDAGData
	for currentBlockGHOSTDAGData.SelectedParent() != nil {
		selectedParentBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, currentBlockGHOSTDAGData.SelectedParent())
		if err != nil {
			return nil, err
		}

		if selectedParentBlockGHOSTDAGData.BlueScore() < blueScore {
			break
		}
		currentHash = currentBlockGHOSTDAGData.SelectedParent()
		currentBlockGHOSTDAGData = selectedParentBlockGHOSTDAGData
	}

	return currentHash, nil
}
