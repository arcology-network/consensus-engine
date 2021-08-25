package types

import (
	tmproto "github.com/HPISTechnologies/consensus-engine/proto/tendermint/types"
)

var ProposalSignBytes func(chainID string, p *tmproto.Proposal) []byte

func init() {
	ProposalSignBytes = DefaultProposalSignBytes
	QuickHash = func(txs Txs) ([]byte, error) {
		return []byte("1234567890abcdefghijklmnopqrstuv"), nil
	}
}
