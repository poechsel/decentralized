package lib

import (
	"sync"
)

type chunkEntry struct {
	chunks []string
}

type FileManager struct {
	lock        *sync.Mutex
	fileToUid   map[string]string
	uidToChunks map[string](map[string]bool)
}

func NewFileManager() *FileManager {
	return &FileManager{
		fileToUid:   make(map[string]string),
		uidToChunks: make(map[string](map[string]bool)),
		lock:        &sync.Mutex{},
	}
}

func (fm *FileManager) AddFile(name string, hash string) {
	fm.lock.Lock()
	defer fm.lock.Unlock()
	if _, ok := fm.uidToChunks[hash]; !ok {
		fm.uidToChunks[hash] = make(map[string]bool)
	}
	fm.fileToUid[name] = hash
}

func (fm *FileManager) AddChunk(hash_file string, hash_chunk string) {
	fm.lock.Lock()
	defer fm.lock.Unlock()
	if _, ok := fm.uidToChunks[hash_file]; !ok {
		fm.uidToChunks[hash_file] = make(map[string]bool)
	}
	fm.uidToChunks[hash_file][hash_chunk] = true
}

func (fm *FileManager) GetChunksFromFilename(filename string) []string {
	fm.lock.Lock()
	defer fm.lock.Unlock()
	if hash, ok := fm.fileToUid[filename]; ok {
		chunks := fm.uidToChunks[hash]
		out := []string{}
		for k := range chunks {
			out = append(out, k)
		}
		return out
	} else {
		return []string{}
	}
}
