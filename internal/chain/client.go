package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client wraps op-geth ethclient for our SDK.
// It provides a simplified interface for chain connectivity.
type Client struct {
	eth     *ethclient.Client
	chainID *big.Int
	wsURL   string
}

// ClientConfig holds configuration for the chain client.
type ClientConfig struct {
	// WSURL is the WebSocket RPC URL (required for subscriptions)
	WSURL string

	// ChainID is the expected chain ID for validation (optional, 0 = skip validation)
	ChainID int64
}

// NewClient creates a new chain client connected via WebSocket.
// Returns error if connection fails or chain ID validation fails.
func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	if cfg.WSURL == "" {
		return nil, fmt.Errorf("chain: WebSocket URL is required")
	}

	// Connect to node via WebSocket
	eth, err := ethclient.DialContext(ctx, cfg.WSURL)
	if err != nil {
		return nil, fmt.Errorf("chain: failed to connect to %s: %w", cfg.WSURL, err)
	}

	// Get chain ID from node
	chainID, err := eth.ChainID(ctx)
	if err != nil {
		eth.Close()
		return nil, fmt.Errorf("chain: failed to get chain ID: %w", err)
	}

	// Validate chain ID if specified
	if cfg.ChainID != 0 && chainID.Int64() != cfg.ChainID {
		eth.Close()
		return nil, fmt.Errorf("chain: chain ID mismatch: expected %d, got %d",
			cfg.ChainID, chainID.Int64())
	}

	return &Client{
		eth:     eth,
		chainID: chainID,
		wsURL:   cfg.WSURL,
	}, nil
}

// ChainID returns the connected chain's ID.
func (c *Client) ChainID() *big.Int {
	return c.chainID
}

// ChainIDUint64 returns the chain ID as uint64.
func (c *Client) ChainIDUint64() uint64 {
	return c.chainID.Uint64()
}

// SubscribeLogs subscribes to log events matching the filter.
// Returns a channel of RawLog and a subscription handle for cancellation.
func (c *Client) SubscribeLogs(ctx context.Context, filter SubscriptionFilter) (<-chan RawLog, ethereum.Subscription, error) {
	// Create go-ethereum filter query
	query := ethereum.FilterQuery{
		Addresses: filter.Addresses,
		Topics:    filter.Topics,
	}

	// Create channel for raw go-ethereum logs
	ethLogs := make(chan types.Log)

	// Subscribe to logs
	sub, err := c.eth.SubscribeFilterLogs(ctx, query, ethLogs)
	if err != nil {
		return nil, nil, fmt.Errorf("chain: subscribe failed: %w", err)
	}

	// Create output channel for our RawLog type
	rawLogs := make(chan RawLog, 100) // Buffer to prevent blocking

	// Start goroutine to convert and forward logs
	go func() {
		defer close(rawLogs)
		for {
			select {
			case <-ctx.Done():
				return
			case log, ok := <-ethLogs:
				if !ok {
					return
				}
				// Convert and send
				select {
				case rawLogs <- FromEthLog(&log):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return rawLogs, sub, nil
}

// GetTransaction fetches a transaction by hash.
func (c *Client) GetTransaction(ctx context.Context, hash common.Hash) (*RawTransaction, error) {
	tx, isPending, err := c.eth.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("chain: get tx failed: %w", err)
	}

	if isPending {
		return nil, fmt.Errorf("chain: transaction is pending")
	}

	// Get sender using the signer
	signer := types.LatestSignerForChainID(c.chainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		return nil, fmt.Errorf("chain: get sender failed: %w", err)
	}

	// Get receipt for block information
	receipt, err := c.eth.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("chain: get receipt failed: %w", err)
	}

	return &RawTransaction{
		Hash:        tx.Hash(),
		Nonce:       tx.Nonce(),
		GasPrice:    tx.GasPrice(),
		GasLimit:    tx.Gas(),
		To:          tx.To(),
		Value:       tx.Value(),
		Data:        tx.Data(),
		From:        from,
		BlockNumber: receipt.BlockNumber.Uint64(),
		BlockHash:   receipt.BlockHash,
		TxIndex:     receipt.TransactionIndex,
	}, nil
}

// GetTransactionReceipt fetches a transaction receipt by hash.
func (c *Client) GetTransactionReceipt(ctx context.Context, hash common.Hash) (*RawReceipt, error) {
	receipt, err := c.eth.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("chain: get receipt failed: %w", err)
	}

	return &RawReceipt{
		Status:      receipt.Status,
		GasUsed:     receipt.GasUsed,
		Logs:        receipt.Logs,
		BlockNumber: receipt.BlockNumber.Uint64(),
		BlockHash:   receipt.BlockHash,
		TxHash:      receipt.TxHash,
	}, nil
}

// GetBlockByNumber fetches block header by number.
// Use nil for latest block.
func (c *Client) GetBlockByNumber(ctx context.Context, number *big.Int) (*RawBlock, error) {
	header, err := c.eth.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("chain: get block failed: %w", err)
	}

	return &RawBlock{
		Number:     header.Number,
		Hash:       header.Hash(),
		ParentHash: header.ParentHash,
		Timestamp:  header.Time,
		GasLimit:   header.GasLimit,
		GasUsed:    header.GasUsed,
		BaseFee:    header.BaseFee,
		Miner:      header.Coinbase,
	}, nil
}

// LatestBlockNumber returns the latest block number.
func (c *Client) LatestBlockNumber(ctx context.Context) (uint64, error) {
	return c.eth.BlockNumber(ctx)
}

// SubscribeNewHead subscribes to new block headers.
// This is supported by more RPC providers than SubscribeFilterLogs.
func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return c.eth.SubscribeNewHead(ctx, ch)
}

// GetBlockLogs fetches all logs for a specific block.
func (c *Client) GetBlockLogs(ctx context.Context, blockNumber *big.Int) ([]RawLog, error) {
	query := ethereum.FilterQuery{
		FromBlock: blockNumber,
		ToBlock:   blockNumber,
	}

	logs, err := c.eth.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("chain: filter logs failed: %w", err)
	}

	rawLogs := make([]RawLog, len(logs))
	for i, log := range logs {
		rawLogs[i] = FromEthLog(&log)
	}

	return rawLogs, nil
}

// Close closes the client connection.
func (c *Client) Close() {
	if c.eth != nil {
		c.eth.Close()
	}
}
