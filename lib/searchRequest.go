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

func (msg *SearchReply) OnReception(state *State, _ func(*GossipPacket)) {
	state.searchRequestCacher.DispatchSearchReply(msg)
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

type OpenedSearch struct {
	pattern             *regexp.Regexp
	transmissionChannel chan (*SearchResultFrom)
}

type SearchRequestCacher struct {
	lock          *sync.Mutex
	cache         map[searchEntry]time.Time
	openedSearch  map[int]OpenedSearch
	uidOpenSearch int
}

func (rc *SearchRequestCacher) CanTreat(request *SearchRequest) bool {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	now := time.Now()
	entry := searchEntry{
		pattern: strings.Join(request.Keywords, ","),
		origin:  request.Origin,
	}

	if time, ok := rc.cache[entry]; ok {
		if now.Sub(time).Seconds() < 0.5 {
			return false
		}
	} else {
		rc.cache[entry] = now
	}
	return true
}

/* Returns the uid of the search */
func (rc *SearchRequestCacher) OpenSearch(keywords []string, outChan chan (*SearchResultFrom)) int {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	uid := rc.uidOpenSearch

	rc.openedSearch[uid] = OpenedSearch{
		pattern:             SearchPatternToRegex(strings.Join(keywords, ",")),
		transmissionChannel: outChan,
	}

	rc.uidOpenSearch += 1

	return uid
}

func (rc *SearchRequestCacher) CloseSearch(uid int) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if _, ok := rc.openedSearch[uid]; ok {
		delete(rc.openedSearch, uid)
	}
}

func (rc *SearchRequestCacher) DispatchSearchReply(r *SearchReply) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	for _, search := range rc.openedSearch {
		for _, result := range r.Results {
			if search.pattern.MatchString(result.FileName) {
				search.transmissionChannel <- &SearchResultFrom{From: r.GetOrigin(), Result: result}
			}
		}
	}
}

func NewSearchRequestCacher() *SearchRequestCacher {
	return &SearchRequestCacher{
		lock:          &sync.Mutex{},
		cache:         make(map[searchEntry]time.Time),
		uidOpenSearch: 0,
		openedSearch:  make(map[int]OpenedSearch),
	}
}

type SearchResultFrom struct {
	From   string
	Result *SearchResult
}

func (sr *SearchResultFrom) String() string {
	chunks_str := []string{}
	for _, c := range sr.Result.ChunkMap {
		chunks_str = append(chunks_str, string(c))
	}
	chunks := strings.Join(chunks_str, ",")
	return "FOUND match " + sr.Result.FileName + " at " + sr.From + " metafile=" + HashToUid(sr.Result.MetafileHash) + " chunks=" + chunks
}
