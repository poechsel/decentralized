package lib

import ()

type searchMergerKey struct {
	name     string
	metaHash string
}

type searchMergerEntry struct {
	count                    uint64
	chunks                   map[uint64](map[string]bool)
	previousNumberDuplicates int
}

type SearchMerger struct {
	content map[searchMergerKey]*searchMergerEntry
}

type SearchAnswer struct {
	FileName string
	MetaHash string
	Chunks   map[uint64]([]string)
}

func (s *SearchAnswer) ToKey() searchMergerKey {
	return searchMergerKey{name: s.FileName, metaHash: s.MetaHash}
}

func NewSearchMerger() *SearchMerger {
	return &SearchMerger{content: make(map[searchMergerKey]*searchMergerEntry)}
}

func (sm *SearchMerger) mergeResult(from string, result *SearchResult) (*SearchAnswer, int) {
	key := searchMergerKey{name: result.FileName, metaHash: HashToUid(result.MetafileHash)}
	if _, err := sm.content[key]; !err {
		sm.content[key] = &searchMergerEntry{
			count:                    result.ChunkCount,
			chunks:                   make(map[uint64](map[string]bool)),
			previousNumberDuplicates: 0,
		}
	}
	for _, chunkId := range result.ChunkMap {
		if _, ok := sm.content[key].chunks[chunkId]; !ok {
			sm.content[key].chunks[chunkId] = make(map[string]bool)
		}
		sm.content[key].chunks[chunkId][from] = true
	}

	if len(sm.content[key].chunks) == int(sm.content[key].count) {
		// first initialize minOcc
		var minOcc int
		for _, e := range sm.content[key].chunks {
			minOcc = len(e)
			break
		}
		// then compute minOcc
		for _, e := range sm.content[key].chunks {
			if len(e) < minOcc {
				minOcc = len(e)
			}
		}

		if minOcc > sm.content[key].previousNumberDuplicates {
			sm.content[key].previousNumberDuplicates = minOcc
			chunks := make(map[uint64]([]string))
			for chunkId, entry := range sm.content[key].chunks {
				chunks[chunkId] = []string{}
				for e, _ := range entry {
					chunks[chunkId] = append(chunks[chunkId], e)
				}
			}
			return &SearchAnswer{
					FileName: key.name,
					MetaHash: key.metaHash,
					Chunks:   chunks,
				},
				minOcc
		}
	}

	return nil, 0
}
