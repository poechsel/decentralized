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

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
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

type SearchRequestCacher struct {
	lock      *sync.Mutex
	timestamp map[searchEntry]time.Time
}

func (rc *SearchRequestCacher) CanTreat(request *SearchRequest) bool {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	now := time.Now()
	entry := searchEntry{
		pattern: strings.Join(request.Keywords, ","),
		origin:  request.Origin,
	}

	if last_seen, ok := rc.timestamp[entry]; ok {
		if now.Sub(last_seen).Seconds() < 0.5 {
			return false
		}
	}
	rc.timestamp[entry] = now
	return true
}

func NewSearchRequestCacher() *SearchRequestCacher {
	return &SearchRequestCacher{
		lock:      &sync.Mutex{},
		timestamp: make(map[searchEntry]time.Time),
	}
}
