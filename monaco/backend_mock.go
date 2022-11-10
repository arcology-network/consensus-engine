package monaco

import (
	"time"

	"github.com/arcology-network/consensus-engine/libs/log"
)

type BackendMock struct {
	logger log.Logger
	txs    [][]byte
	ch     chan [][]byte
}

func NewBackendMock() *BackendMock {
	return &BackendMock{
		ch: make(chan [][]byte, 1000),
	}
}

func (bm *BackendMock) SetLogger(logger log.Logger) {
	bm.logger = logger
}

func (bm *BackendMock) Reap(int64, int64) ([][]byte, [][]byte) {
	return [][]byte{
			{1, 2, 3},
		}, [][]byte{
			{4, 5, 6},
		}
}

func (bm *BackendMock) AddToMempool(txs [][]byte, _ string) {
	bm.logger.Debug("AddToMempool", "txs", txs)
	bm.txs = txs
}

func (bm *BackendMock) ApplyTxsSync(height int64, coinbase []byte, timestamp time.Time, hashes [][]byte) []byte {
	bm.logger.Debug("ApplyTxsSync", "height", height, "coinbase", coinbase, "txs", bm.txs, "hashes", hashes)
	return []byte{7, 8, 9}
}

func (bm *BackendMock) GetLocalTxsChan() chan [][]byte {
	return bm.ch
}

func (bm *BackendMock) GetTxsOnBlock(height uint64) ([][]byte, error) {
	return nil, nil
}
