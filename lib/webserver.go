package lib

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
	"time"
)

type WebServer struct {
	server            *http.Server
	AddPeerChannel    chan string
	AddMessageChannel chan Message
	messages_lock     *sync.RWMutex
	messages          []Message
	peers_lock        *sync.RWMutex
	peers             []string
	nameServer        string
}

func (websrv *WebServer) ListenEvents() {
	for {
		select {
		case peer := <-websrv.AddPeerChannel:
			websrv.peers_lock.Lock()
			websrv.peers = append(websrv.peers, peer)
			websrv.peers_lock.Unlock()
		case msg := <-websrv.AddMessageChannel:
			websrv.messages_lock.Lock()
			websrv.messages = append(websrv.messages, msg)
			websrv.messages_lock.Unlock()
		}
	}
}

func NewWebServer(state *State, server *Gossiper, address string) *WebServer {
	r := mux.NewRouter()

	name := server.Name

	srv := &http.Server{
		Handler: r,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	websrv := &WebServer{server: srv,
		nameServer:        name,
		messages:          []Message{},
		peers:             []string{},
		AddPeerChannel:    make(chan string, 64),
		AddMessageChannel: make(chan Message, 64),
		messages_lock:     &sync.RWMutex{},
		peers_lock:        &sync.RWMutex{},
	}

	go websrv.ListenEvents()

	type ServerId struct {
		Name    string
		Address string
	}

	r.HandleFunc("/node",
		func(_ http.ResponseWriter, r *http.Request) {
			var peer string
			json.NewDecoder(r.Body).Decode(&peer)
			state.AddPeer(peer)
		}).Methods("POST")

	r.HandleFunc("/message",
		func(_ http.ResponseWriter, r *http.Request) {
			var message string
			json.NewDecoder(r.Body).Decode(&message)
			rumor := RumorMessage{
				Origin: server.Name,
				ID:     server.NewMsgId(),
				Text:   message}
			server.HandleRumor(state, server.Address.String(), &rumor)
		}).Methods("POST")

	r.HandleFunc("/id",
		func(w http.ResponseWriter, _ *http.Request) {
			e := ServerId{Name: server.Name, Address: server.Address.String()}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(e)
		}).Methods("GET")

	r.HandleFunc("/node",
		func(w http.ResponseWriter, _ *http.Request) {
			websrv.peers_lock.RLock()
			defer websrv.peers_lock.RUnlock()
			json.NewEncoder(w).Encode(websrv.peers)
		}).Methods("GET")

	r.HandleFunc("/message",
		func(w http.ResponseWriter, _ *http.Request) {
			websrv.messages_lock.Lock()
			defer websrv.messages_lock.Unlock()
			json.NewEncoder(w).Encode(websrv.messages)
			websrv.messages = []Message{}
		}).Methods("GET")

	/* we also serve a bunch of static files */
	r.PathPrefix("/").Handler(
		http.StripPrefix("/", http.FileServer(http.Dir("./gui/dist"))))

	return websrv
}

func (srv *WebServer) Start() error {
	return srv.server.ListenAndServe()
}
