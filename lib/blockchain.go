package lib

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type BlockChainNode struct {
	parent   [32]byte
	children [][32]byte
	block    *Block
}

func NewBlockChainNode(block *Block) *BlockChainNode {
	return &BlockChainNode{
		parent:   block.PrevHash,
		children: [][32]byte{},
		block:    block,
	}
}

type MineEndSignal struct {
	block    Block
	duration time.Duration
}

type BlockChainMapEntry struct {
	hash [32]byte
	nb   int
}

func NewBlockChainMapEntry(hash [32]byte) *BlockChainMapEntry {
	return &BlockChainMapEntry{
		hash: hash,
		nb:   0,
	}
}

type BlockChain struct {
	nameToHash     map[string]*BlockChainMapEntry
	blocks         map[[32]byte](*BlockChainNode)
	nextFilesToAdd []TxPublish
	isMining       bool
	headChain      [32]byte
	// Sometime, we can get a Block A with parent B before
	// receiving Block B. We store somewhere that we now that
	// B has a child, so to set a correct information when creating
	// it. We store the list of pending children for this node
	waitingToBeResolved map[[32]byte][][32]byte

	blockMinedSignal chan MineEndSignal

	ReleaseBlock chan Block
	AddBlock     chan Block
	TryBlock     chan TryWrapper
	AddTxPublish chan TxPublish
	TryTxPublish chan TryWrapper
}

func (blockchain *BlockChain) FindForkHelper(start [32]byte, seen map[[32]byte]bool, conv map[[32]byte]string) {
	if _, ok := seen[start]; ok {
	} else {
		seen[start] = true
		if _, ok := blockchain.blocks[start]; !ok || IsZeroHash(start[:]) {
		} else {
			blockchain.FindForkHelper(blockchain.blocks[start].parent, seen, conv)
		}
	}
}

func (blockchain *BlockChain) ReverseTransactionsChain(start [32]byte, step int, seen map[[32]byte]bool, conv map[[32]byte]string) (int, [32]byte) {
	if _, ok := seen[start]; ok {
		return step, start
	} else {
		seen[start] = true
		if _, ok := blockchain.blocks[start]; !ok || IsZeroHash(start[:]) {
			return step, start
		} else {
			//fmt.Println("reverting ", conv[start])
			blockchain.ReverseTransaction(blockchain.blocks[start].block)
			return blockchain.ReverseTransactionsChain(blockchain.blocks[start].parent, step+1, seen, conv)
		}
	}
}
func (blockchain *BlockChain) ApplyTransactionsChain(start [32]byte, until [32]byte, conv map[[32]byte]string) {
	if _, ok := blockchain.blocks[start]; !ok || IsZeroHash(start[:]) || start == until {
	} else {
		//fmt.Println("applying ", conv[start])
		blockchain.ApplyTransaction(blockchain.blocks[start].block)
		blockchain.ApplyTransactionsChain(blockchain.blocks[start].parent, until, conv)
	}
}

func (BlockChain *BlockChain) Dump(conv map[[32]byte]string) {
	for hash, node := range BlockChain.blocks {
		fmt.Println(conv[hash], "->", conv[node.parent], len(node.children))
	}
}

/** Compute the length of the chain ending at [start]
Sometime we can be missing a node (for ex for the chain a--b--[c]--d--e if
we suppose we haven't received [c] when asking for the length of the chain
ending at e). In this case we return a negative length **/
func (BlockChain *BlockChain) ComputeLengthChain(start [32]byte) int {
	if IsZeroHash(start[:]) {
		return 0
	} else if _, ok := BlockChain.blocks[start]; !ok {
		return -(int(^uint(0) >> 1)) - 1
	} else {
		return BlockChain.ComputeLengthChain(BlockChain.blocks[start].parent) + 1
	}
}

func (blockchain *BlockChain) ApplyTransaction(block *Block) {
	for _, t := range block.Transactions {
		if _, ok := blockchain.nameToHash[t.File.Name]; !ok {
			hash := [32]byte{}
			/*for i, c := range t.File.MetafileHash {
				if i < len(hash) {
					hash[i] = c
				}
			}
			*/
			copy(hash[:], t.File.MetafileHash)
			blockchain.nameToHash[t.File.Name] = NewBlockChainMapEntry(hash)
		}
		blockchain.nameToHash[t.File.Name].nb += 1
	}
}

func (blockchain *BlockChain) ReverseTransaction(block *Block) {
	for _, t := range block.Transactions {
		if _, ok := blockchain.nameToHash[t.File.Name]; ok {
			blockchain.nameToHash[t.File.Name].nb -= 1
			if blockchain.nameToHash[t.File.Name].nb <= 0 {
				delete(blockchain.nameToHash, t.File.Name)
			}
		}
	}
}

func (blockchain *BlockChain) getLeaves(start [32]byte) [][32]byte {
	if block, ok := blockchain.blocks[start]; ok {
		if len(block.children) == 0 {
			return [][32]byte{start}
		} else {
			o := [][32]byte{}
			for _, child := range block.children {
				o = append(o, blockchain.getLeaves(child)...)
			}
			return o
		}
	} else {
		return [][32]byte{}
	}
}

func (blockchain *BlockChain) UpdateLongestChain(conv map[[32]byte]string, hash [32]byte) {
	if blockchain.ComputeLengthChain(hash) > blockchain.ComputeLengthChain(blockchain.headChain) {
		seen := make(map[[32]byte]bool)
		blockchain.FindForkHelper(hash, seen, conv)
		//	fmt.Println()
		rewind, stop := blockchain.ReverseTransactionsChain(blockchain.headChain, 0, seen, conv)
		//	fmt.Println()
		//	fmt.Println(rewind)
		blockchain.ApplyTransactionsChain(hash, stop, conv)
		block := blockchain.blocks[hash].block
		if block.PrevHash != blockchain.headChain {
			fmt.Println("FORK-LONGER", "rewind", rewind, "blocks")
			log.Println("FORK-LONGER", "rewind", rewind, "blocks")
		}
		blockchain.headChain = hash
	}

}

func (blockchain *BlockChain) AppendBlock(conv map[[32]byte]string, block *Block) {

	isForkShorter := false
	hash := block.Hash()

	//fmt.Println("\n\n\nInserting ", conv[block.Hash()], "head =", conv[blockchain.headChain])
	//fmt.Println(HashToUid(hash[:]), "<>", HashToUid(blockchain.headChain[:]))

	if parentBcn, ok := blockchain.blocks[block.PrevHash]; (ok && len(parentBcn.children) > 0) ||
		(!ok && len(blockchain.waitingToBeResolved[block.PrevHash]) > 0) {
		isForkShorter = true
	}
	if parentBcn, ok := blockchain.blocks[block.PrevHash]; ok {
		parentBcn.children = append(parentBcn.children, hash)
	} else {
		blockchain.waitingToBeResolved[block.PrevHash] = append(blockchain.waitingToBeResolved[block.PrevHash], hash)
	}

	if isForkShorter {
		fmt.Println("FORK-SHORTER", HashToUid(hash[:]))
		log.Println("FORK-SHORTER", HashToUid(hash[:]))
	}

	currentBcn := NewBlockChainNode(block)
	blockchain.blocks[hash] = currentBcn

	//fmt.Println(blockchain.ComputeLengthChain(hash), blockchain.ComputeLengthChain(blockchain.headChain))

	blockchain.UpdateLongestChain(conv, hash)

	// if we already no this node has a child, then we fix the flag
	// This is like "missing node found"
	//Why doing it as the end ? Because we want to make sure every transactions are applied
	// Suppose we missed node C and we just saw it.
	// Then, if we are under: A-B-[C]-D-E, by first updating the longest chain with C we will
	// make sure every transactions before C are done, and then we can do securely the transactions
	// between E and C
	// Otherwise, we are only gonna do the transactions between E and C, C excluded. And it could
	// be even worse if before C the longest chain was A-I-J-K, and hence the transaction on
	// B wasn't yet computed. At least, that's how I see it

	if entry, ok := blockchain.waitingToBeResolved[hash]; ok {
		currentBcn.children = entry
		delete(blockchain.waitingToBeResolved, hash)
		leaves := blockchain.getLeaves(hash)
		log.Println("saw missing node", leaves)
		for _, leaf := range leaves {
			blockchain.UpdateLongestChain(conv, leaf)
		}
	}

	//fmt.Println("new head =", conv[blockchain.headChain])

}

type TryWrapper struct {
	callback chan bool
	content  interface{}
}

func NewTryWrapper(c interface{}) *TryWrapper {
	return &TryWrapper{content: c, callback: make(chan bool)}
}

func (bc *BlockChain) ChainToString(start [32]byte) string {
	if IsZeroHash(start[:]) {
		return ""
	} else {
		if _, ok := bc.blocks[start]; ok {
			block := bc.blocks[start].block
			blockstr := HashToUid(start[:]) + ":"
			blockstr += HashToUid(block.PrevHash[:]) + ":"

			trstr := []string{}
			for _, ct := range block.Transactions {
				trstr = append(trstr, ct.File.Name)
			}

			blockstr += strings.Join(trstr, ",")

			return blockstr + " " + bc.ChainToString(bc.blocks[start].parent)
		} else {
			return ""
		}
	}
}

func IsZeroHash(hash []byte) bool {
	for _, x := range hash {
		if x != 0 {
			return false
		}
	}
	return true
}

func (bc *BlockChain) mineInner(transaction []TxPublish, prevhash [32]byte) {
	nextblock, time := Mine(transaction, prevhash)
	bc.blockMinedSignal <- MineEndSignal{block: *nextblock, duration: time}
}
func (bc *BlockChain) MineNextBlock() {
	transaction := bc.GetNextTransactionsToMine()

	if !bc.isMining /*&& len(transaction) > 0 */ {
		bc.nextFilesToAdd = []TxPublish{}
		bc.isMining = true
		go bc.mineInner(transaction, bc.headChain)
	}
}

func (blockchain *BlockChain) GetNextTransactionsToMine() []TxPublish {
	transaction := []TxPublish{}
	for _, t := range blockchain.nextFilesToAdd {
		if entry, ok := blockchain.nameToHash[t.File.Name]; !ok || (ok && entry.nb <= 0) {
			transaction = append(transaction, t)

		}
	}
	return transaction
}

func (bc *BlockChain) Work() {
	bc.isMining = true
	go bc.mineInner([]TxPublish{}, [32]byte{})
	for {
		select {
		case signal := <-bc.blockMinedSignal:
			block := signal.block
			bc.isMining = false

			bc.AddBlock <- block

			go func(signal MineEndSignal) {
				if IsZeroHash(signal.block.PrevHash[:]) {
					time.Sleep(5 * time.Second)
				} else {
					time.Sleep(2 * signal.duration)
				}
				bc.ReleaseBlock <- signal.block
			}(signal)

		case block := <-bc.AddBlock:
			if _, ok := bc.blocks[block.PrevHash]; (ok || IsZeroHash(block.PrevHash[:])) && block.IsValid() {
				bc.AppendBlock(make(map[[32]byte]string), &block)

				bc.MineNextBlock()
				/*nextfiles := []TxPublish{}

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
				*/

				fmt.Println("CHAIN", bc.ChainToString(bc.headChain))
				//log.Println("CHAIN", bc.ChainToString(bc.headChain))
			}

		case tryBlock := <-bc.TryBlock:
			block := tryBlock.content.(Block)
			_, ok := bc.blocks[block.Hash()]
			tryBlock.callback <- block.IsValid() && !ok

		case tryTxPublish := <-bc.TryTxPublish:
			t := tryTxPublish.content.(TxPublish)
			entry, ok := bc.nameToHash[t.File.Name]
			seen := false
			for _, f := range bc.nextFilesToAdd {
				seen = seen || (f.File.Name == t.File.Name &&
					f.File.Size == t.File.Size &&
					HashToUid(f.File.MetafileHash) == HashToUid(t.File.MetafileHash))
			}

			tryTxPublish.callback <- (!ok || (ok && entry.nb <= 0)) && !seen

		case txPublish := <-bc.AddTxPublish:
			bc.nextFilesToAdd = append(bc.nextFilesToAdd, txPublish)

			bc.MineNextBlock()
		}
	}
}

func NewBlockChain() *BlockChain {
	bc := &BlockChain{
		isMining:            false,
		nameToHash:          make(map[string]*BlockChainMapEntry),
		blocks:              make(map[[32]byte]*BlockChainNode),
		nextFilesToAdd:      []TxPublish{},
		blockMinedSignal:    make(chan MineEndSignal, 10),
		ReleaseBlock:        make(chan Block, 64),
		AddBlock:            make(chan Block, 64),
		TryBlock:            make(chan TryWrapper, 64),
		AddTxPublish:        make(chan TxPublish, 64),
		TryTxPublish:        make(chan TryWrapper, 64),
		waitingToBeResolved: make(map[[32]byte][][32]byte),
	}

	bc.headChain = [32]byte{}
	bcn := &BlockChainNode{parent: bc.headChain, children: [][32]byte{}, block: nil}
	bc.blocks[bc.headChain] = bcn
	return bc
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
	try := NewTryWrapper(msg.Block)
	state.BlockChain.TryBlock <- *try
	select {
	case answer := <-try.callback:
		if answer {
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

func Mine(transactions []TxPublish, PrevHash [32]byte) (*Block, time.Duration) {
	start := time.Now()
	for {
		nonce := [32]byte{}
		rand.Read(nonce[:])
		block := NewBlock(PrevHash, nonce, transactions)
		if block.IsValid() {
			hash := block.Hash()
			fmt.Println("FOUND-BLOCK", HashToUid(hash[:]))
			return block, time.Since(start)
		}
	}
	return nil, time.Since(start)
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
