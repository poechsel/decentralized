package lib

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
)

type Message struct {
	Address string
	Rumor   RumorMessage
}

type DataAckKey struct {
	Hash string
	Peer string
}

/* The State contains all knowledge we have about the world.
This include the messages, the peers...
*/

type State struct {
	lock_peers *sync.RWMutex
	/* map of peer addresses to peer struct*/
	known_peers map[string]*Peer
	/* list of peer addresses */
	list_peers []string
	/* database of all messages */
	db *Database
	/* routing table */
	lock_routing *sync.RWMutex
	routing      map[string]string

	/* Two channels used to notify when we
	are seeing a new message or adding a new peer */
	addMessageChannels        [](chan Message)
	addPrivateMessageChannels [](chan PrivateMessage)
	addPeerChannels           [](chan string)

	lockDataAck *sync.Mutex
	dataAck     map[DataAckKey]Stack

	FileManager         *FileManager
	searchRequestCacher *SearchRequestCacher
	FileKnowledgeDB     *FileKnowledgeDB
}

func (state *State) DispatchDataAck(peer string, hash string, ack DataReply) bool {
	state.lockDataAck.Lock()
	defer state.lockDataAck.Unlock()
	key := DataAckKey{Peer: peer, Hash: hash}
	if s, ok := state.dataAck[key]; ok {
		var has_dispatched bool = false
		for {
			if s.Empty() {
				break
			} else {
				c := s.Pop().(AckRequest)
				if c.SendAck(ack) {
					has_dispatched = true
					break
				}
			}
		}
		if s.Empty() {
			delete(state.dataAck, key)
		}
		return has_dispatched
	} else {
		return false
	}
}

func (state *State) AddDataAck(peer string, hash string, c AckRequest) {
	state.lockDataAck.Lock()
	defer state.lockDataAck.Unlock()
	key := DataAckKey{Peer: peer, Hash: hash}
	if s, ok := state.dataAck[key]; ok {
		s.Push(c)
		state.dataAck[key] = s
	} else {
		s := NewStack()
		s.Push(c)
		state.dataAck[key] = *s
	}
}

func (state *State) GetRoutingTableNames() []string {
	state.lock_routing.RLock()
	defer state.lock_routing.RUnlock()
	out := []string{}
	for key, _ := range state.routing {
		out = append(out, key)
	}
	return out
}

func NewState() *State {
	db := NewDatabase()
	state := &State{
		known_peers:         make(map[string]*Peer),
		db:                  &db,
		lock_peers:          &sync.RWMutex{},
		lock_routing:        &sync.RWMutex{},
		routing:             make(map[string]string),
		lockDataAck:         &sync.Mutex{},
		dataAck:             make(map[DataAckKey]Stack),
		FileManager:         NewFileManager(),
		searchRequestCacher: NewSearchRequestCacher(),
		FileKnowledgeDB:     NewFileKnowledgeDB(),
	}
	return state
}

func (state *State) getRouteTo(peer string) (string, bool) {
	state.lock_routing.RLock()
	defer state.lock_routing.RUnlock()
	route, ok := state.routing[peer]
	return route, ok
}

func (state *State) UpdateRoutingTable(peer string, address string) {
	state.lock_routing.Lock()
	defer state.lock_routing.Unlock()
	fmt.Println("DSDV", peer, address)
	state.routing[peer] = address
}

/* Get a random peer that is not in the list avoir */
func (state *State) getRandomPeer(avoid ...string) (*Peer, error) {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()

	restricted := make(map[string]bool)
	for _, a := range avoid {
		restricted[a] = true
	}
	var peers []string
	for _, name := range state.list_peers {
		if _, ok := restricted[name]; !ok {
			peers = append(peers, name)
		}
	}

	if len(peers) == 0 {
		return nil, errors.New("No peer to select from")
	}
	k := rand.Intn(len(peers))
	name := peers[k]
	return state.known_peers[name], nil
}

/* Get N random peers that are not in the list avoid */
func (state *State) getNRandomPeer(n int, currentPeerAddress string) ([](*Peer), error) {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()

	var peers []string
	for _, name := range state.list_peers {
		if name != currentPeerAddress {
			peers = append(peers, name)
		}
	}

	generated := [](*Peer){}
	if n >= len(peers) {
		dbgnames := []string{}
		for _, p := range peers {
			generated = append(generated, state.known_peers[p])
			dbgnames = append(dbgnames, p)
		}
		log.Println("random -> ", dbgnames)
		return generated, nil
	} else {
		perm := rand.Perm(len(peers))
		dbgnames := []string{}
		for i, k := range perm {
			if i < n {
				dbgnames = append(dbgnames, peers[k])
				generated = append(generated, state.known_peers[peers[k]])
			}
		}
		log.Println("random -> ", dbgnames)
		return generated, nil
	}
}

func (state *State) AddNewPeerCallback(c chan string) {
	state.addPeerChannels = append(state.addPeerChannels, c)
}
func (state *State) AddNewMessageCallback(c chan Message) {
	state.addMessageChannels = append(state.addMessageChannels, c)
}
func (state *State) AddNewPrivateMessageCallback(c chan PrivateMessage) {
	state.addPrivateMessageChannels = append(state.addPrivateMessageChannels, c)
}

/* If the peer at address "address" requests an ack, then dispatch
the statuspacket "status" to him */
func (state *State) dispatchStatusToPeer(address string, status *StatusPacket) bool {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()

	if peer, ok := state.known_peers[address]; ok {
		if peer.DispatchStatus(status) {
			return true
		}
	}
	return false
}

/* Add a peer and notify the channels which subscribed to this event */
func (state *State) AddPeer(address string) bool {
	state.lock_peers.Lock()
	defer state.lock_peers.Unlock()
	if _, ok := state.known_peers[address]; ok || address == "" {
		return false
	} else {
		peer, err := NewPeer(address)
		if err == nil {
			state.known_peers[address] = peer
			state.list_peers = append(state.list_peers, address)
			for _, c := range state.addPeerChannels {
				c <- address
			}
		}
		return true
	}
}

func (state *State) String() string {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()
	keys := make([]string, 0, len(state.known_peers))
	for key := range state.known_peers {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}

/* Execute a function on all peers. The function is not called on
peers having address "avoid" */
func (state *State) IterPeers(avoid string, fct func(*Peer)) {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()
	for addr, peer := range state.known_peers {
		if avoid != addr {
			fct(peer)
		}
	}
}

/* Return a tuple of booleans
The first one is true if we succeeded in adding the rumor message
The second one is true if we have seen a message with id greater than
the current one stored
*/
func (state *State) addRumorMessage(rumor *RumorMessage, sender_addr_string string) (bool, bool) {
	minNotPresent := state.db.GetMinNotPresent(rumor.Origin)
	isIdGreater := rumor.ID >= minNotPresent
	if minNotPresent == rumor.ID && !state.db.PossessRumorMessage(rumor) {
		state.db.InsertRumorMessage(rumor)

		if rumor.Text != "" {
			for _, c := range state.addMessageChannels {
				c <- Message{Rumor: *rumor, Address: sender_addr_string}
			}
		}
		return true, isIdGreater
	} else {
		return false, isIdGreater
	}
}

func (state *State) addPrivateMessage(private *PrivateMessage) {
	for _, c := range state.addPrivateMessageChannels {
		c <- *private
	}
}
