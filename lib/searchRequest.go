package lib

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

func SearchPatternToRegex(pattern string) *regexp.Regexp {
	pattern = strings.Replace(pattern, ".", "\\.", -1)
	pattern = strings.Replace(pattern, "*", ".*", -1)
	pattern = strings.Replace(pattern, "|", "\\|", -1)
	pattern = strings.Replace(pattern, ",", "|", -1)
	pattern = "^(" + pattern + ")$"
	return regexp.MustCompile(pattern)
}

func FilenameMatchPattern(pattern *regexp.Regexp, filename string) bool {
	return pattern.MatchString(filename)
}

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

func NewSearchRequest(origin string, budget uint64, keywords []string) *SearchRequest {
	return &SearchRequest{Origin: origin, Budget: budget, Keywords: keywords}
}

type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

func (msg *SearchReply) ToPacket() *GossipPacket {
	return &GossipPacket{SearchReply: msg}
}

func (msg *SearchReply) GetOrigin() string {
	return msg.Origin
}

func (msg *SearchReply) GetDestination() string {
	return msg.Destination
}

func NewSearchReply(origin string, destination string, results []*SearchResult) *SearchReply {
	o := SearchReply{
		Origin:      origin,
		Destination: destination,
		HopLimit:    10,
		Results:     results,
	}
	return &o
}

func (msg *SearchReply) NextHop() bool {
	msg.HopLimit -= 1
	if msg.HopLimit <= 0 {
		return false
	} else {
		return true
	}
}

func (msg *SearchReply) OnFirstEmission(state *State) {
}

func (msg *SearchReply) OnReception(_ *State, _ func(*GossipPacket)) {
	//TODO do something
}

type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}

type searchEntry struct {
	pattern string
	origin  string
}

type resultEntry struct {
	chunkCount   uint64
	MetafileHash []byte
	ChunkMap     map[uint64]int
}

func (r *resultEntry) numberComplete() int {
	if len(r.ChunkMap) < int(r.chunkCount) {
		return 0
	}

	var minNumber int
	for _, minNumber = range r.ChunkMap {
		break
	}

	for _, n := range r.ChunkMap {
		if n < minNumber {
			minNumber = n
		}
	}
	return minNumber
}

func newResultEntry(metafile []byte, count uint64) *resultEntry {
	return &resultEntry{
		chunkCount:   count,
		MetafileHash: metafile,
		ChunkMap:     make(map[uint64]int),
	}
}

type resultEntryKey struct {
	name     string
	metahash string
}

type searchResultEntry struct {
	time       time.Time
	pattern_re *regexp.Regexp
	results    map[resultEntryKey](*resultEntry)
}

func (r *searchResultEntry) numberComplete() int {
	var o int = 0
	for _, r := range r.results {
		o += r.numberComplete()
	}
	return o
}

type SearchRequestCacher struct {
	lock  *sync.Mutex
	cache map[searchEntry](*searchResultEntry)
}

func newSearchResultEntry(time time.Time, pattern string) *searchResultEntry {
	return &searchResultEntry{
		time:       time,
		results:    make(map[resultEntryKey](*resultEntry)),
		pattern_re: SearchPatternToRegex(pattern),
	}
}

func (rc *SearchRequestCacher) AppendSearchResult(result *SearchResult) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	key := resultEntryKey{name: result.FileName, metahash: HashToUid((result.MetafileHash))}

	for _, entry := range rc.cache {
		if FilenameMatchPattern(entry.pattern_re, result.FileName) {
			if _, ok := entry.results[key]; !ok {
				entry.results[key] = newResultEntry(result.MetafileHash, result.ChunkCount)
			}
			entry_file := entry.results[key]
			for _, chunk := range result.ChunkMap {
				// when looking on a non present entry it returns 0 by default
				entry_file.ChunkMap[chunk] = entry_file.ChunkMap[chunk] + 1
			}
		}
	}
}

func (rc *SearchRequestCacher) CanTreat(request *SearchRequest) bool {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	now := time.Now()
	entry := searchEntry{
		pattern: strings.Join(request.Keywords, ","),
		origin:  request.Origin,
	}

	if content, ok := rc.cache[entry]; ok {
		if now.Sub(content.time).Seconds() < 0.5 {
			return false
		}
	} else {
		rc.cache[entry] = newSearchResultEntry(now, entry.pattern)
	}
	rc.cache[entry].time = now
	return true
}

func NewSearchRequestCacher() *SearchRequestCacher {
	return &SearchRequestCacher{
		lock:  &sync.Mutex{},
		cache: make(map[searchEntry](*searchResultEntry)),
	}
}
