package monaco

import "time"

type BackendProxy interface {
	Reap(maxBytes int64, maxGas int64) (txs [][]byte, hashes [][]byte)
	AddToMempool(txs [][]byte)
	ApplyTxsSync(height int64, coinbase []byte, timestamp time.Time, hashes [][]byte) []byte
	GetLocalTxsChan() chan [][]byte
}
