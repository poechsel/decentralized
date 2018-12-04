package lib

import (
	"regexp"
	"sync"
)

type chunkEntry struct {
	chunks []string
}

type metaFileInformation struct {
	hash  string
	count uint64
}

type FileManager struct {
	lock        *sync.Mutex
	fileToUid   map[string]metaFileInformation
	uidToChunks map[string](map[string]uint64)
}

func NewFileManager() *FileManager {
	return &FileManager{
		fileToUid:   make(map[string]metaFileInformation),
		uidToChunks: make(map[string](map[string]uint64)),
		lock:        &sync.Mutex{},
	}
}

func (fm *FileManager) toSearchReply(pattern *regexp.Regexp) [](*SearchResult) {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	out := [](*SearchResult){}

	for fileName, metafile := range fm.fileToUid {
		if FilenameMatchPattern(pattern, fileName) {
			chunkMap := []uint64{}
			for _, pos := range fm.uidToChunks[metafile.hash] {
				chunkMap = append(chunkMap, pos)
			}
			result := &SearchResult{
				FileName:     fileName,
				MetafileHash: UidToHash(metafile.hash),
				ChunkMap:     chunkMap,
				ChunkCount:   metafile.count,
			}
			out = append(out, result)
		}
	}
	return out
}

func (fm *FileManager) AddFile(name string, hash string, count uint64) {
	fm.lock.Lock()
	defer fm.lock.Unlock()
	if _, ok := fm.uidToChunks[hash]; !ok {
		fm.uidToChunks[hash] = make(map[string]uint64)
	}
	fm.fileToUid[name] = metaFileInformation{hash: hash, count: count}
}

func (fm *FileManager) AddChunk(hash_file string, hash_chunk string, chunkPos uint64) {
	fm.lock.Lock()
	defer fm.lock.Unlock()
	if _, ok := fm.uidToChunks[hash_file]; !ok {
		fm.uidToChunks[hash_file] = make(map[string]uint64)
	}
	fm.uidToChunks[hash_file][hash_chunk] = chunkPos
}

func (fm *FileManager) GetChunksFromFilename(filename string) []string {
	fm.lock.Lock()
	defer fm.lock.Unlock()
	if metaFile, ok := fm.fileToUid[filename]; ok {
		chunks := fm.uidToChunks[metaFile.hash]
		out := []string{}
		for k := range chunks {
			out = append(out, k)
		}
		return out
	} else {
		return []string{}
	}
}
