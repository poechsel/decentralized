package lib

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

type Message struct {
	Address string
	Rumor   RumorMessage
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
}

func (state *State) GetRoutingTable() map[string]string {
	state.lock_routing.RLock()
	defer state.lock_routing.RUnlock()
	out := make(map[string]string)
	for key, value := range state.routing {
		out[key] = value
	}
	return out
}

func NewState() *State {
	db := NewDatabase()
	state := &State{
		known_peers:  make(map[string]*Peer),
		db:           &db,
		lock_peers:   &sync.RWMutex{},
		lock_routing: &sync.RWMutex{},
		routing:      make(map[string]string),
	}
	return state
}

func (state *State) getRouteTo(peer string) (string, bool) {
	state.lock_routing.RLock()
	defer state.lock_routing.RUnlock()
	route, ok := state.routing[peer]
	return route, ok
}

func (state *State) updateRoutingTable(peer string, address string) {
	state.lock_routing.Lock()
	defer state.lock_routing.Unlock()
	fmt.Println("DSDV", peer, address)
	state.routing[peer] = address
}

/* Get a random peer that is not in the list avoir */
func (state *State) getRandomPeer(avoid ...string) (string, *Peer, error) {
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
		return "", nil, errors.New("No peer to select from")
	}
	k := rand.Intn(len(peers))
	name := peers[k]
	return name, state.known_peers[name], nil
}

func (state *State) AddNewPeerCallback(c chan string) {
	state.addPeerChannels = append(state.addPeerChannels, c)
}
func (state *State) AddNewMessageCallback(c chan Message) {
	state.addMessageChannels = append(state.addMessageChannels, c)
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
	if !state.db.PossessRumorMessage(rumor) &&
		minNotPresent == rumor.ID {

		if rumor.Text != "" {
			state.db.InsertRumorMessage(rumor)
		}

		for _, c := range state.addMessageChannels {
			c <- Message{Rumor: *rumor, Address: sender_addr_string}
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
