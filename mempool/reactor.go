package mempool

import (
	"errors"
	"fmt"
	"math"
	"time"

	cfg "github.com/arcology-network/consensus-engine/config"
	"github.com/arcology-network/consensus-engine/libs/log"
	tmsync "github.com/arcology-network/consensus-engine/libs/sync"
	"github.com/arcology-network/consensus-engine/monaco"
	"github.com/arcology-network/consensus-engine/p2p"
	protomem "github.com/arcology-network/consensus-engine/proto/tendermint/mempool"
	"github.com/arcology-network/consensus-engine/types"
)

const (
	MempoolChannel = byte(0x30)

	peerCatchupSleepIntervalMS = 100 // If peer is behind, sleep this amount

	// UnknownPeerID is the peer ID to use when running CheckTx when there is
	// no peer (e.g. RPC)
	UnknownPeerID uint16 = 0

	maxActiveIDs = math.MaxUint16
)

// Reactor handles mempool tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping txs to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	config  *cfg.MempoolConfig
	mempool *CListMempool
	ids     *mempoolIDs
	backend monaco.BackendProxy
}

type mempoolIDs struct {
	mtx       tmsync.RWMutex
	peerMap   map[p2p.ID]uint16
	nextID    uint16              // assumes that a node will never have over 65536 active peers
	activeIDs map[uint16]struct{} // used to check if a given peerID key is used, the value doesn't matter
	chanMap   map[p2p.ID]chan [][]byte
}

// Reserve searches for the next unused ID and assigns it to the
// peer.
func (ids *mempoolIDs) ReserveForPeer(peer p2p.Peer) {
	ids.mtx.Lock()
	defer ids.mtx.Unlock()

	curID := ids.nextPeerID()
	ids.peerMap[peer.ID()] = curID
	ids.activeIDs[curID] = struct{}{}
	ids.chanMap[peer.ID()] = make(chan [][]byte, 1000)
}

// nextPeerID returns the next unused peer ID to use.
// This assumes that ids's mutex is already locked.
func (ids *mempoolIDs) nextPeerID() uint16 {
	if len(ids.activeIDs) == maxActiveIDs {
		panic(fmt.Sprintf("node has maximum %d active IDs and wanted to get one more", maxActiveIDs))
	}

	_, idExists := ids.activeIDs[ids.nextID]
	for idExists {
		ids.nextID++
		_, idExists = ids.activeIDs[ids.nextID]
	}
	curID := ids.nextID
	ids.nextID++
	return curID
}

// Reclaim returns the ID reserved for the peer back to unused pool.
func (ids *mempoolIDs) Reclaim(peer p2p.Peer) {
	ids.mtx.Lock()
	defer ids.mtx.Unlock()

	removedID, ok := ids.peerMap[peer.ID()]
	if ok {
		delete(ids.activeIDs, removedID)
		delete(ids.peerMap, peer.ID())
		delete(ids.chanMap, peer.ID())
	}
}

// GetForPeer returns an ID reserved for the peer.
func (ids *mempoolIDs) GetForPeer(peer p2p.Peer) uint16 {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	return ids.peerMap[peer.ID()]
}

// GetForPeerEx used only in Monaco.
func (ids *mempoolIDs) GetForPeerEx(peer p2p.Peer) (uint16, chan [][]byte) {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	return ids.peerMap[peer.ID()], ids.chanMap[peer.ID()]
}

func newMempoolIDs() *mempoolIDs {
	return &mempoolIDs{
		peerMap:   make(map[p2p.ID]uint16),
		activeIDs: map[uint16]struct{}{0: {}},
		nextID:    1, // reserve unknownPeerID(0) for mempoolReactor.BroadcastTx
		chanMap:   make(map[p2p.ID]chan [][]byte),
	}
}

// NewReactor returns a new Reactor with the given config and mempool.
func NewReactor(config *cfg.MempoolConfig, mempool *CListMempool) *Reactor {
	memR := &Reactor{
		config:  config,
		mempool: mempool,
		ids:     newMempoolIDs(),
	}
	memR.BaseReactor = *p2p.NewBaseReactor("Mempool", memR)
	return memR
}

func (memR *Reactor) SetBackendProxy(backend monaco.BackendProxy) {
	memR.backend = backend
}

// InitPeer implements Reactor by creating a state for the peer.
func (memR *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
	memR.ids.ReserveForPeer(peer)
	return peer
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (memR *Reactor) SetLogger(l log.Logger) {
	memR.Logger = l
	memR.mempool.SetLogger(l)
}

// OnStart implements p2p.BaseReactor.
func (memR *Reactor) OnStart() error {
	if !memR.config.Broadcast {
		memR.Logger.Info("Tx broadcasting is disabled")
	}

	ch := memR.backend.GetLocalTxsChan()
	go func() {
		for {
			if !memR.IsRunning() {
				return
			}

			select {
			case txs := <-ch:
				memR.ids.mtx.Lock()
				for _, c := range memR.ids.chanMap {
					c <- txs
				}
				memR.ids.mtx.Unlock()
			case <-memR.Quit():
				return
			default:
				time.Sleep(30 * time.Millisecond)
			}
		}
	}()
	return nil
}

// GetChannels implements Reactor by returning the list of channels for this
// reactor.
func (memR *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	largestTx := make([]byte, memR.config.MaxTxBytes)
	batchMsg := protomem.Message{
		Sum: &protomem.Message_Txs{
			Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
		},
	}

	return []*p2p.ChannelDescriptor{
		{
			ID:                  MempoolChannel,
			Priority:            5,
			SendQueueCapacity:   1024,
			RecvMessageCapacity: batchMsg.Size(),
		},
	}
}

// AddPeer implements Reactor.
// It starts a broadcast routine ensuring all txs are forwarded to the given peer.
func (memR *Reactor) AddPeer(peer p2p.Peer) {
	if memR.config.Broadcast {
		go memR.broadcastTxRoutine(peer)
	}
}

// RemovePeer implements Reactor.
func (memR *Reactor) RemovePeer(peer p2p.Peer, reason interface{}) {
	memR.ids.Reclaim(peer)
	// broadcast routine checks if peer is gone and returns
}

// Receive implements Reactor.
// It adds any received transactions to the mempool.
func (memR *Reactor) Receive(chID byte, src p2p.Peer, msgBytes []byte) {
	msg, err := memR.decodeMsg(msgBytes)
	if err != nil {
		memR.Logger.Error("Error decoding message", "src", src, "chId", chID, "err", err)
		memR.Switch.StopPeerForError(src, err)
		return
	}
	// memR.Logger.Debug("Receive", "src", src, "chId", chID, "msg", msg)

	txInfo := TxInfo{SenderID: memR.ids.GetForPeer(src)}
	if src != nil {
		txInfo.SenderP2PID = src.ID()
	}

	txs := make([][]byte, len(msg.Txs))
	for i := range txs {
		txs[i] = msg.Txs[i]
	}
	memR.backend.AddToMempool(txs, string(src.ID()))
	// broadcasting happens from go routines per peer
}

// PeerState describes the state of a peer.
type PeerState interface {
	GetHeight() int64
}

// Send new mempool txs to peer.
func (memR *Reactor) broadcastTxRoutine(peer p2p.Peer) {
	_, ch := memR.ids.GetForPeerEx(peer)

	for {
		// In case of both next.NextWaitChan() and peer.Quit() are variable at the same time
		if !memR.IsRunning() || !peer.IsRunning() {
			// if ch != nil {
			// 	close(ch)
			// }
			return
		}

		select {
		case txs := <-ch:
			msg := protomem.Message{
				Sum: &protomem.Message_Txs{
					Txs: &protomem.Txs{Txs: txs},
				},
			}
			bz, err := msg.Marshal()
			if err != nil {
				panic(err)
			}
			success := peer.Send(MempoolChannel, bz)
			if !success {
				time.Sleep(peerCatchupSleepIntervalMS * time.Millisecond)
				continue
			}
		case <-peer.Quit():
			return
		case <-memR.Quit():
			return
		default:
			time.Sleep(30 * time.Millisecond)
		}
	}
}

//-----------------------------------------------------------------------------
// Messages

func (memR *Reactor) decodeMsg(bz []byte) (TxsMessage, error) {
	msg := protomem.Message{}
	err := msg.Unmarshal(bz)
	if err != nil {
		return TxsMessage{}, err
	}

	var message TxsMessage

	if i, ok := msg.Sum.(*protomem.Message_Txs); ok {
		txs := i.Txs.GetTxs()

		if len(txs) == 0 {
			return message, errors.New("empty TxsMessage")
		}

		decoded := make([]types.Tx, len(txs))
		for j, tx := range txs {
			decoded[j] = types.Tx(tx)
		}

		message = TxsMessage{
			Txs: decoded,
		}
		return message, nil
	}
	return message, fmt.Errorf("msg type: %T is not supported", msg)
}

//-------------------------------------

// TxsMessage is a Message containing transactions.
type TxsMessage struct {
	Txs []types.Tx
}

// String returns a string representation of the TxsMessage.
func (m *TxsMessage) String() string {
	return fmt.Sprintf("[TxsMessage %v]", m.Txs)
}
