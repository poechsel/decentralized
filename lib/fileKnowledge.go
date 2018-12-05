package lib

import (
	"math/rand"
	"sync"
)

/* Manage chunks known to be stored on distant (including our) node */

type FileKnowledgeInfo struct {
	peersHavingMetahash    []string
	peersHavingChunk       map[int]([]string)
	peersHavingMetahashMap map[string]bool
	peersHavingChunkMap    map[int](map[string]bool)
}

type FileKnowledgeDB struct {
	lock *sync.Mutex
	/* map of metafile -> chunk_id -> map of peers */
	db map[string]*FileKnowledgeInfo
}

func NewFileKnowledgeInfo() *FileKnowledgeInfo {
	return &FileKnowledgeInfo{
		peersHavingChunk:       make(map[int][]string),
		peersHavingMetahash:    []string{},
		peersHavingChunkMap:    make(map[int](map[string]bool)),
		peersHavingMetahashMap: make(map[string]bool),
	}
}

func (k *FileKnowledgeInfo) Insert(chunkId int, peer string) {
	if _, ok := k.peersHavingMetahashMap[peer]; !ok {
		k.peersHavingMetahashMap[peer] = true
		k.peersHavingMetahash = append(k.peersHavingMetahash, peer)
	}
	if _, ok := k.peersHavingChunkMap[chunkId]; !ok {
		k.peersHavingChunkMap[chunkId] = make(map[string]bool)
		k.peersHavingChunk[chunkId] = []string{}
	}
	if _, ok := k.peersHavingChunkMap[chunkId][peer]; !ok {
		k.peersHavingChunkMap[chunkId][peer] = true
		k.peersHavingChunk[chunkId] = append(k.peersHavingChunk[chunkId], peer)
	}
}

func NewFileKnowledgeDB() *FileKnowledgeDB {
	return &FileKnowledgeDB{db: make(map[string]*FileKnowledgeInfo)}
}

func (fk *FileKnowledgeDB) Insert(metahash string, chunkId int, peer string) {
	fk.lock.Lock()
	defer fk.lock.Unlock()

	if _, ok := fk.db[metahash]; !ok {
		fk.db[metahash] = NewFileKnowledgeInfo()
	}

	fk.db[metahash].Insert(chunkId, peer)
}

func (fk *FileKnowledgeDB) SelectPeerForMetaHash(peer string, metahash string) string {
	if peer != "" {
		return peer
	}
	fk.lock.Lock()
	defer fk.lock.Unlock()

	if entry, ok := fk.db[metahash]; ok {
		return entry.peersHavingMetahash[rand.Intn(len(entry.peersHavingMetahash))]
	} else {
		return peer
	}
}

func (fk *FileKnowledgeDB) SelectPeerForChunk(peer string, metahash string, chunkId int) string {
	if peer != "" {
		return peer
	}
	fk.lock.Lock()
	defer fk.lock.Unlock()

	if entryM, ok := fk.db[metahash]; ok {
		if entry, ok := entryM.peersHavingChunk[chunkId]; ok {
			return entry[rand.Intn(len(entry))]

		}
	}
	return peer
}
