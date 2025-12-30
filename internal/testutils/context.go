package testutils

import (
	"github.com/asentric/asentric/pkg/domain"
)

type MockContext struct {
	chainID domain.ChainID
	tx      *domain.Transaction
	block   *domain.Block
	logs    []domain.Log
	abi     domain.ABIRegistry
}

func NewMockContext() *MockContext {
	return &MockContext{
		chainID: domain.ChainID(1),
		tx:      defaultTransaction(),
		block:   defaultBlock(),
		logs:    defaultLogs(),
		abi:     nil,
	}
}

func (m *MockContext) ChainID() domain.ChainID {
	return m.chainID
}

func (m *MockContext) Tx() domain.Transaction {
	if m.tx == nil {
		return domain.Transaction{}
	}
	return *m.tx
}

func (m *MockContext) Block() domain.Block {
	if m.block == nil {
		return domain.Block{}
	}
	return *m.block
}

func (m *MockContext) Logs() []domain.Log {
	return m.logs
}

func (m *MockContext) ABI() domain.ABIRegistry {
	return m.abi
}

func (m *MockContext) WithChainID(chainID domain.ChainID) *MockContext {
	m.chainID = chainID
	return m
}

// WithTransaction sets the transaction.
func (m *MockContext) WithTransaction(tx *domain.Transaction) *MockContext {
	m.tx = tx
	return m
}

// WithBlock sets the block.
func (m *MockContext) WithBlock(block *domain.Block) *MockContext {
	m.block = block
	return m
}

// WithLogs sets the logs.
func (m *MockContext) WithLogs(logs []domain.Log) *MockContext {
	m.logs = logs
	return m
}

// WithABI sets the ABI registry.
func (m *MockContext) WithABI(abi domain.ABIRegistry) *MockContext {
	m.abi = abi
	return m
}

// defaultTransaction creates a default transaction for testing.
func defaultTransaction() *domain.Transaction {
	return &domain.Transaction{
		Hash:         domain.Hash("0xdefault123456789abcdef"),
		Index:        0,
		From:         domain.Address("0xSender123"),
		To:           domain.Address("0xReceiver456"),
		Nonce:        1,
		GasLimit:     21000,
		GasUsed:      21000,
		Status:       true,
		RawValue:     domain.NativeValue{Wei: "1000000000000000000"},
		GasPrice:     "20000000000",
		MaxFeePerGas: "0",
		MaxPriority:  "0",
		Type:         domain.TxTypeDynamicFee,
		Action:       domain.TxActionCall,
		Call:         nil,
		BlockNumber:  1000000,
		BlockHash:    domain.Hash("0xblock123"),
		Timestamp:    1703500800,
	}
}

// defaultBlock creates a default block for testing.
func defaultBlock() *domain.Block {
	return &domain.Block{
		Hash:      domain.Hash("0xblock123"),
		Parent:    domain.Hash("0xparent456"),
		Number:    1000000,
		Timestamp: 1703500800, // Dec 25, 2024
		Miner:     domain.Address("0xMiner789"),
	}
}

// defaultLogs creates default logs for testing.
func defaultLogs() []domain.Log {
	return []domain.Log{
		{
			Address:     domain.Address("0xContract123"),
			LogIndex:    0,
			TxHash:      domain.Hash("0xdefault123456789abcdef"),
			TxIndex:     0,
			Event:       domain.Event{Name: "Transfer", Fields: map[string]any{"from": "0xSender123", "to": "0xReceiver456", "amount": "1000000000000000000"}},
			BlockNumber: 1000000,
			BlockHash:   domain.Hash("0xblock123"),
		},
	}
}
