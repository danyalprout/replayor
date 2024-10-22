package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockEngineAPI struct {
	server *httptest.Server
}

func NewMockEngineAPI(t *testing.T) *MockEngineAPI {
	mock := &MockEngineAPI{}
	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		switch req["method"] {
		case "engine_newPayloadV1":
			mock.handleNewPayload(w)
		case "engine_forkchoiceUpdatedV1":
			mock.handleForkChoiceUpdated(w)
		default:
			http.Error(w, "Unsupported method", http.StatusNotImplemented)
		}
	}))

	return mock
}

func (m *MockEngineAPI) URL() string {
	return m.server.URL
}

func (m *MockEngineAPI) Close() {
	m.server.Close()
}

func (m *MockEngineAPI) handleNewPayload(w http.ResponseWriter) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"status":          "VALID",
			"latestValidHash": hexutil.Encode([]byte{0x01, 0x23, 0x45, 0x67}),
			"validationError": nil,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockEngineAPI) handleForkChoiceUpdated(w http.ResponseWriter) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"payloadStatus": map[string]interface{}{
				"status":          "VALID",
				"latestValidHash": hexutil.Encode([]byte{0x01, 0x23, 0x45, 0x67}),
				"validationError": nil,
			},
			"payloadId": nil,
		},
	}
	json.NewEncoder(w).Encode(response)
}

type MockEthNode struct {
	server *httptest.Server
}

func NewMockEthNode(t *testing.T) *MockEthNode {
	mock := &MockEthNode{}
	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		switch req["method"] {
		case "eth_blockNumber":
			mock.handleBlockNumber(w)
		case "eth_getBlockByNumber":
			mock.handleGetBlockByNumber(w)
		default:
			http.Error(w, "Unsupported method", http.StatusNotImplemented)
		}
	}))

	return mock
}

func (m *MockEthNode) URL() string {
	return m.server.URL
}

func (m *MockEthNode) Close() {
	m.server.Close()
}

func (m *MockEthNode) handleBlockNumber(w http.ResponseWriter) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result":  "0x1234",
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockEthNode) handleGetBlockByNumber(w http.ResponseWriter) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"number":           "0x1234",
			"hash":             "0x0000000000000000000000000000000000000000000000000000000000000000",
			"parentHash":       "0x0000000000000000000000000000000000000000000000000000000000000000",
			"nonce":            "0x0000000000000000",
			"mixHash":          "0x0000000000000000000000000000000000000000000000000000000000000000",
			"sha3Uncles":       "0x0000000000000000000000000000000000000000000000000000000000000000",
			"logsBloom":        "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			"transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
			"stateRoot":        "0x0000000000000000000000000000000000000000000000000000000000000000",
			"receiptsRoot":     "0x0000000000000000000000000000000000000000000000000000000000000000",
			"miner":            "0x0000000000000000000000000000000000000000",
			"difficulty":       "0x0",
			"totalDifficulty":  "0x0",
			"extraData":        "0x",
			"size":             "0x0",
			"gasLimit":         "0x0",
			"gasUsed":          "0x0",
			"timestamp":        "0x0",
			"transactions":     []interface{}{},
			"uncles":           []interface{}{},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func TestSetupClients(t *testing.T) {
	mockEngineAPI := NewMockEngineAPI(t)
	defer mockEngineAPI.Close()

	mockSourceNode := NewMockEthNode(t)
	defer mockSourceNode.Close()

	mockDestNode := NewMockEthNode(t)
	defer mockDestNode.Close()

	// Mock config
	cfg := config.ReplayorConfig{
		SourceNodeUrl:   mockSourceNode.URL(),
		ExecutionUrl:    mockDestNode.URL(),
		EngineApiUrl:    mockEngineAPI.URL(),
		EngineApiSecret: common.BytesToHash([]byte("test-secret")),
		RollupConfig:    &rollup.Config{}, // Add necessary fields
	}

	logger := log.New()
	ctx := context.Background()

	// Test successful setup
	t.Run("Successful Setup", func(t *testing.T) {
		clients, err := SetupClients(cfg, logger, ctx)
		require.NoError(t, err)
		assert.NotNil(t, clients.SourceNode)
		assert.NotNil(t, clients.DestNode)
		assert.NotNil(t, clients.EngineApi)

	})

	// Test connection retry
	t.Run("Connection Retry", func(t *testing.T) {
		// Modify cfg to use unavailable endpoints
		cfg.ExecutionUrl = "http://unavailable-execution:8545"
		cfg.EngineApiUrl = "http://unavailable-engine:8551"
		cfg.SourceNodeUrl = "http://unavailable-source:8545"

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := SetupClients(cfg, logger, ctx)
		assert.Error(t, err)
	})

}
