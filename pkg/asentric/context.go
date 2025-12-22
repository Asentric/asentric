package asentric

import (
	"github.com/asentric/asentric/pkg/domain"
)

type Context interface {
	ChainID() domain.ChainID
	Tx() domain.Transaction
	Block() domain.Block
	Logs() []domain.Log
	ABI() domain.ABIRegistry
}
