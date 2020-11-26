package externalapi

import (
	"bytes"
	"fmt"
)

// DomainTransaction represents a Kaspa transaction
type DomainTransaction struct {
	Version      int32
	Inputs       []*DomainTransactionInput
	Outputs      []*DomainTransactionOutput
	LockTime     uint64
	SubnetworkID DomainSubnetworkID
	Gas          uint64
	PayloadHash  DomainHash
	Payload      []byte

	Fee  uint64
	Mass uint64
}

// Clone returns a clone of DomainTransaction
func (tx *DomainTransaction) Clone() *DomainTransaction {
	if tx == nil {
		return nil
	}

	payloadClone := make([]byte, len(tx.Payload))
	copy(payloadClone, tx.Payload)

	inputsClone := make([]*DomainTransactionInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		inputsClone[i] = input.Clone()
	}

	outputsClone := make([]*DomainTransactionOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		outputsClone[i] = output.Clone()
	}

	return &DomainTransaction{
		Version:      tx.Version,
		Inputs:       inputsClone,
		Outputs:      outputsClone,
		LockTime:     tx.LockTime,
		SubnetworkID: *tx.SubnetworkID.Clone(),
		Gas:          tx.Gas,
		PayloadHash:  *tx.PayloadHash.Clone(),
		Payload:      payloadClone,
		Fee:          tx.Fee,
		Mass:         tx.Mass,
	}
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = DomainTransaction{0, []*DomainTransactionInput{}, []*DomainTransactionOutput{}, 0,
	DomainSubnetworkID{}, 0, DomainHash{}, []byte{}, 0, 0}

// Equal returns whether tx equals to other
func (tx *DomainTransaction) Equal(other *DomainTransaction) bool {
	if tx == nil || other == nil {
		return tx == other
	}

	if tx.Version != other.Version {
		return false
	}

	for i, input := range tx.Inputs {
		if !input.Equal(other.Inputs[i]) {
			return false
		}
	}

	for i, output := range tx.Outputs {
		if !output.Equal(other.Outputs[i]) {
			return false
		}
	}

	if tx.LockTime != other.LockTime {
		return false
	}

	if !tx.SubnetworkID.Equal(&other.SubnetworkID) {
		return false
	}

	if tx.Gas != other.Gas {
		return false
	}

	if !tx.PayloadHash.Equal(&other.PayloadHash) {
		return false
	}

	if !bytes.Equal(tx.Payload, other.Payload) {
		return false
	}

	if tx.Fee != other.Fee {
		return false
	}

	if tx.Mass != other.Mass {
		return false
	}

	return true
}

// DomainTransactionInput represents a Kaspa transaction input
type DomainTransactionInput struct {
	PreviousOutpoint DomainOutpoint
	SignatureScript  []byte
	Sequence         uint64

	UTXOEntry *UTXOEntry
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = &DomainTransactionInput{DomainOutpoint{}, []byte{}, 0,
	&UTXOEntry{}}

// Equal returns whether input equals to other
func (input *DomainTransactionInput) Equal(other *DomainTransactionInput) bool {
	if input == nil || other == nil {
		return input == other
	}

	if !input.PreviousOutpoint.Equal(&other.PreviousOutpoint) {
		return false
	}

	if !bytes.Equal(input.SignatureScript, other.SignatureScript) {
		return false
	}

	if input.Sequence != other.Sequence {
		return false
	}

	if !input.UTXOEntry.Equal(other.UTXOEntry) {
		return false
	}

	return true
}

// Clone returns a clone of DomainTransactionInput
func (input *DomainTransactionInput) Clone() *DomainTransactionInput {
	if input == nil {
		return nil
	}

	signatureScriptClone := make([]byte, len(input.SignatureScript))
	copy(signatureScriptClone, input.SignatureScript)

	return &DomainTransactionInput{
		PreviousOutpoint: *input.PreviousOutpoint.Clone(),
		SignatureScript:  signatureScriptClone,
		Sequence:         input.Sequence,
		UTXOEntry:        input.UTXOEntry.Clone(),
	}
}

// DomainOutpoint represents a Kaspa transaction outpoint
type DomainOutpoint struct {
	TransactionID DomainTransactionID
	Index         uint32
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = DomainOutpoint{DomainTransactionID{}, 0}

// Equal returns whether op equals to other
func (op *DomainOutpoint) Equal(other *DomainOutpoint) bool {
	if op == nil || other == nil {
		return op == other
	}

	return *op == *other
}

// Clone returns a clone of DomainOutpoint
func (op *DomainOutpoint) Clone() *DomainOutpoint {
	if op == nil {
		return nil
	}

	return &DomainOutpoint{
		TransactionID: *op.TransactionID.Clone(),
		Index:         op.Index,
	}
}

// String stringifies an outpoint.
func (op DomainOutpoint) String() string {
	return fmt.Sprintf("(%s: %d)", op.TransactionID, op.Index)
}

// NewDomainOutpoint instantiates a new DomainOutpoint with the given id and index
func NewDomainOutpoint(id *DomainTransactionID, index uint32) *DomainOutpoint {
	return &DomainOutpoint{
		TransactionID: *id,
		Index:         index,
	}
}

// DomainTransactionOutput represents a Kaspad transaction output
type DomainTransactionOutput struct {
	Value           uint64
	ScriptPublicKey []byte
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = DomainTransactionOutput{0, []byte{}}

// Equal returns whether output equals to other
func (output *DomainTransactionOutput) Equal(other *DomainTransactionOutput) bool {
	if output == nil || other == nil {
		return output == other
	}

	if output.Value != other.Value {
		return false
	}

	if !bytes.Equal(output.ScriptPublicKey, other.ScriptPublicKey) {
		return false
	}

	return true
}

// Clone returns a clone of DomainTransactionOutput
func (output *DomainTransactionOutput) Clone() *DomainTransactionOutput {
	if output == nil {
		return nil
	}

	scriptPublicKeyClone := make([]byte, len(output.ScriptPublicKey))
	copy(scriptPublicKeyClone, output.ScriptPublicKey)

	return &DomainTransactionOutput{
		Value:           output.Value,
		ScriptPublicKey: scriptPublicKeyClone,
	}
}

// DomainTransactionID represents the ID of a Kaspa transaction
type DomainTransactionID DomainHash

// String stringifies a transaction ID.
func (id DomainTransactionID) String() string {
	return DomainHash(id).String()
}

// Clone returns a clone of DomainTransactionID
func (id *DomainTransactionID) Clone() *DomainTransactionID {
	if id == nil {
		return nil
	}

	idClone := *id
	return &idClone
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ DomainTransactionID = [DomainHashSize]byte{}

// Equal returns whether id equals to other
func (id *DomainTransactionID) Equal(other *DomainTransactionID) bool {
	if id == nil || other == nil {
		return id == other
	}

	return *id == *other
}
