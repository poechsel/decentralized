package lib

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
)

type BlockChain struct {
	nameToHash     map[string][]byte
	allBlocks      [](*Block)
	mapBlocks      map[[32]byte](*Block)
	nextFilesToAdd []TxPublish
	isMining       bool
	prevHash       [32]byte

	blockMinedSignal chan Block

	ReleaseBlock chan Block
	AddBlock     chan Block
	TryBlock     chan TryWrapper
	AddTxPublish chan TxPublish
	TryTxPublish chan TryWrapper
}

type TryWrapper struct {
	callback chan bool
	content  interface{}
}

func NewTryWrapper(c interface{}) *TryWrapper {
	return &TryWrapper{content: c, callback: make(chan bool)}
}

func (bc *BlockChain) ShowChain() {
	s := ""
	for i := len(bc.allBlocks) - 1; i >= 0; i-- {
		block := bc.allBlocks[i]
		hash := block.Hash()
		blockstr := HashToUid(hash[:]) + ":"
		blockstr += HashToUid(block.PrevHash[:]) + ":"

		trstr := []string{}
		for _, ct := range block.Transactions {
			trstr = append(trstr, ct.File.Name)
		}

		blockstr += strings.Join(trstr, ",")

		s += "[" + blockstr + "]" + " "
	}

	fmt.Println("CHAIN", s)
	log.Println("CHAIN", s)
}

func IsZeroHash(hash []byte) bool {
	for _, x := range hash {
		if x != 0 {
			return false
		}
	}
	return true
}

func (bc *BlockChain) Mine(transaction []TxPublish, prevhash [32]byte) {
	nextblock := Mine(transaction, prevhash)
	bc.blockMinedSignal <- *nextblock
}

func (bc *BlockChain) Work() {
	for {
		select {
		case block := <-bc.blockMinedSignal:
			bc.isMining = false
			transaction := make([]TxPublish, len(bc.nextFilesToAdd))
			copy(transaction, bc.nextFilesToAdd)
			bc.nextFilesToAdd = []TxPublish{}

			if len(transaction) > 0 {
				bc.isMining = true
				go bc.Mine(transaction, block.Hash())
			}

			bc.AddBlock <- block
			bc.ReleaseBlock <- block

		case block := <-bc.AddBlock:
			log.Println("testing add", block.PrevHash, IsZeroHash(block.PrevHash[:]))
			if _, ok := bc.mapBlocks[block.PrevHash]; (ok || IsZeroHash(block.PrevHash[:])) && block.IsValid() {
				bc.allBlocks = append(bc.allBlocks, &block)
				bc.prevHash = block.Hash()
				bc.mapBlocks[block.Hash()] = &block

				nextfiles := []TxPublish{}

				for _, t := range block.Transactions {
					bc.nameToHash[t.File.Name] = t.File.MetafileHash
					seen := false
					for _, f := range block.Transactions {
						seen = seen || (f.File.Name == t.File.Name &&
							f.File.Size == t.File.Size &&
							HashToUid(f.File.MetafileHash) == HashToUid(t.File.MetafileHash))
					}
					if !seen {
						nextfiles = append(nextfiles, t)
					}
				}
				bc.nextFilesToAdd = nextfiles

				bc.ShowChain()
			}

		case tryBlock := <-bc.TryBlock:
			block := tryBlock.content.(Block)
			_, ok := bc.mapBlocks[block.Hash()]
			tryBlock.callback <- block.IsValid() && !ok

		case tryTxPublish := <-bc.TryTxPublish:
			t := tryTxPublish.content.(TxPublish)
			_, ok := bc.nameToHash[t.File.Name]
			seen := false
			for _, f := range bc.nextFilesToAdd {
				seen = seen || (f.File.Name == t.File.Name &&
					f.File.Size == t.File.Size &&
					HashToUid(f.File.MetafileHash) == HashToUid(t.File.MetafileHash))
			}

			tryTxPublish.callback <- !ok && !seen

		case txPublish := <-bc.AddTxPublish:
			bc.nextFilesToAdd = append(bc.nextFilesToAdd, txPublish)

			if bc.isMining == false {
				bc.isMining = true
				go bc.Mine(bc.nextFilesToAdd, bc.prevHash)
				bc.nextFilesToAdd = []TxPublish{}
			}
		}
	}
}

func NewBlockChain() *BlockChain {
	return &BlockChain{
		isMining:         false,
		nameToHash:       make(map[string]([]byte)),
		allBlocks:        []*Block{},
		mapBlocks:        make(map[[32]byte]*Block),
		nextFilesToAdd:   []TxPublish{},
		blockMinedSignal: make(chan Block, 10),
		ReleaseBlock:     make(chan Block, 64),
		AddBlock:         make(chan Block, 64),
		TryBlock:         make(chan TryWrapper, 64),
		AddTxPublish:     make(chan TxPublish, 64),
		TryTxPublish:     make(chan TryWrapper, 64),
		prevHash:         [32]byte{},
	}

}

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
	/* return false if we can't consider the message
	If we can consider it, then will also act on it */
	IsValidAndReceive(state *State) bool
	ToPacket() *GossipPacket
}

type TxPublish struct {
	File     File
	HopLimit uint32
}

func NewTxPublish(name string, metafilehash []byte, filesize int64) TxPublish {
	return TxPublish{
		File: File{Name: name,
			MetafileHash: metafilehash,
			Size:         filesize},
		HopLimit: 10,
	}
}

func (msg *TxPublish) IsValidAndReceive(state *State) bool {
	var txpublish TxPublish
	txpublish = *msg
	try := NewTryWrapper(txpublish)
	state.BlockChain.TryTxPublish <- *try
	select {
	case answer := <-try.callback:
		if answer {
			state.BlockChain.AddTxPublish <- txpublish
		}
		return answer
	}
	return false
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

func NewBlockPublish(block Block) *BlockPublish {
	return &BlockPublish{Block: block, HopLimit: 20}
}

func (msg *BlockPublish) NextHop() (BroadcastWithLimit, bool) {
	log.Println("#####", "ok")
	if msg.HopLimit <= 1 {
		return msg, false
	} else {
		return &BlockPublish{
			Block:    msg.Block,
			HopLimit: msg.HopLimit - 1,
		}, true
	}
}
func (msg *BlockPublish) IsValidAndReceive(state *State) bool {
	log.Println("##", msg)
	try := NewTryWrapper(msg.Block)
	state.BlockChain.TryBlock <- *try
	select {
	case answer := <-try.callback:
		log.Println("---->", answer)
		if answer {
			log.Println("#########", msg)
			state.BlockChain.AddBlock <- msg.Block
		}
		return answer
	}
	return false
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

func (b *Block) IsValid() bool {
	hash := b.Hash()
	isValid := true
	/* Because there are 8bit in a byte */
	for i := 0; i < 16/8; i++ {
		isValid = isValid && (hash[i] == 0)
	}
	return isValid
}

func Mine(transactions []TxPublish, PrevHash [32]byte) *Block {
	for {
		nonce := [32]byte{}
		rand.Read(nonce[:])
		block := NewBlock(PrevHash, nonce, transactions)
		if block.IsValid() {
			hash := block.Hash()
			fmt.Println("FOUND-BLOCK", "["+HashToUid(hash[:])+"]")
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
