package lib

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
)

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

	/* Two channels used to notify when we
	are seeing a new message or adding a new peer */
	addMessageChannels [](chan Message)
	addPeerChannels    [](chan string)
}

func NewState() *State {
	db := NewDatabase()
	state := &State{
		known_peers: make(map[string]*Peer),
		db:          &db,
		lock_peers:  &sync.RWMutex{}}
	return state
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

type Message struct {
	Address string
	Rumor   RumorMessage
}

/* Return true if we succeeded in adding the rumor message */
func (state *State) addRumorMessage(rumor *RumorMessage, sender_addr_string string) bool {
	if !state.db.PossessRumorMessage(rumor) &&
		state.db.GetMinNotPresent(rumor.Origin) == rumor.ID {

		state.db.InsertRumorMessage(rumor)

		for _, c := range state.addMessageChannels {
			c <- Message{Rumor: *rumor, Address: sender_addr_string}
		}
		return true
	} else {
		return false
	}
}
