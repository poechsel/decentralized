package lib

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
)

type BroadcastWithLimitCacher struct {
	lock  *sync.Mutex
	cache map[[32]byte]bool
}

func NewBroadcastWithLimitCacher() *BroadcastWithLimitCacher {
	return &BroadcastWithLimitCacher{
		lock:  &sync.Mutex{},
		cache: make(map[[32]byte]bool),
	}
}

func (b *BroadcastWithLimitCacher) CanTreat(bw BroadcastWithLimit) bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	key := bw.ToKey()
	if _, ok := b.cache[key]; ok {
		return false
	} else {
		b.cache[key] = true
		return true
	}
}

type BroadcastWithLimit interface {
	NextHop() (BroadcastWithLimit, bool)
	ToKey() [32]byte
	IsValid(state *State) bool
	Received(state *State)
	ToPacket() *GossipPacket
}

type TxPublish struct {
	File     File
	HopLimit uint32
}

func (msg *TxPublish) IsValid(state *State) bool {
	// TODO
	return true
}

func (msg *TxPublish) Received(state *State) {

}

func (msg *TxPublish) ToPacket() *GossipPacket {
	return &GossipPacket{TxPublish: msg}
}

func (msg *TxPublish) NextHop() (BroadcastWithLimit, bool) {
	if msg.HopLimit <= 1 {
		return msg, false
	} else {
		return &TxPublish{
			File:     msg.File,
			HopLimit: msg.HopLimit - 1,
		}, true
	}
}

func (msg *TxPublish) ToKey() [32]byte {
	return msg.Hash()
}

func (msg *BlockPublish) ToKey() [32]byte {
	return msg.Block.Hash()
}

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

func (msg *BlockPublish) NextHop() (BroadcastWithLimit, bool) {
	if msg.HopLimit <= 1 {
		return msg, false
	} else {
		return &BlockPublish{
			Block:    msg.Block,
			HopLimit: msg.HopLimit - 1,
		}, true
	}
}
func (msg *BlockPublish) IsValid(state *State) bool {
	// TODO
	return true
}

func (msg *BlockPublish) Received(state *State) {

}

func (msg *BlockPublish) ToPacket() *GossipPacket {
	return &GossipPacket{BlockPublish: msg}
}

type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

func NewBlock(prev [32]byte, nonce [32]byte, transactions []TxPublish) *Block {
	return &Block{
		PrevHash:     prev,
		Nonce:        nonce,
		Transactions: transactions,
	}
}

func Mine(transactions []TxPublish, PrevHash [32]byte) *Block {
	for {
		nonce := [32]byte{}
		rand.Read(nonce[:])
		block := NewBlock(PrevHash, nonce, transactions)
		hash := block.Hash()
		isValid := true
		for i := 0; i < 16; i++ {
			isValid = isValid && (hash[i] != 0)
		}
		if isValid {
			fmt.Println("FOUND-BLOCK", HashToUid(hash[:]))
			return block
		}
	}
	return nil
}

func (b *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])
	h.Write(b.Nonce[:])
	binary.Write(h, binary.LittleEndian,
		uint32(len(b.Transactions)))
	for _, t := range b.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}
func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian,
		uint32(len(t.File.Name)))
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}
