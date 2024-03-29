package lib

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PrivatePost struct {
	To      string
	Content string
}

type FileRequest struct {
	Peer      string
	HashValue string
	Filename  string
}

type WebSearchResult struct {
	FileName string
	Keywords string
	MetaHash string
}

type WebServer struct {
	server                   *http.Server
	AddPeerChannel           chan string
	AddMessageChannel        chan Message
	AddPrivateMessageChannel chan PrivateMessage
	AddSearchResultChannel   chan WebSearchResult
	messages_lock            *sync.RWMutex
	messages                 []Message
	searchresults_lock       *sync.RWMutex
	searchresults            []WebSearchResult
	peers_lock               *sync.RWMutex
	peers                    []string

	private_lock *sync.RWMutex
	private      []PrivateMessage
	nameServer   string
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
		case msg := <-websrv.AddSearchResultChannel:
			websrv.searchresults_lock.Lock()
			websrv.searchresults = append(websrv.searchresults, msg)
			websrv.searchresults_lock.Unlock()
		case msg := <-websrv.AddPrivateMessageChannel:
			websrv.private_lock.Lock()
			websrv.private = append(websrv.private, msg)
			websrv.private_lock.Unlock()
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
		nameServer:               name,
		messages:                 []Message{},
		peers:                    []string{},
		searchresults:            []WebSearchResult{},
		AddPeerChannel:           make(chan string, 64),
		AddMessageChannel:        make(chan Message, 64),
		AddPrivateMessageChannel: make(chan PrivateMessage, 64),
		AddSearchResultChannel:   make(chan WebSearchResult, 64),
		messages_lock:            &sync.RWMutex{},
		peers_lock:               &sync.RWMutex{},
		private_lock:             &sync.RWMutex{},
		searchresults_lock:       &sync.RWMutex{},
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

	r.HandleFunc("/private",
		func(_ http.ResponseWriter, r *http.Request) {
			var message PrivatePost
			json.NewDecoder(r.Body).Decode(&message)
			private := NewPrivateMessage(websrv.nameServer, message.Content, message.To)
			server.HandlePointToPointMessage(state, server.Address.String(), &private)
		}).Methods("POST")

	r.HandleFunc("/upload",
		func(_ http.ResponseWriter, r *http.Request) {
			var message string
			json.NewDecoder(r.Body).Decode(&message)
			go server.UploadFile(state, message)
		}).Methods("POST")

	r.HandleFunc("/search",
		func(_ http.ResponseWriter, r *http.Request) {
			var keywords string
			json.NewDecoder(r.Body).Decode(&keywords)
			go server.LaunchSearch(state, strings.Split(keywords, ","), 0)
		}).Methods("POST")

	r.HandleFunc("/download",
		func(_ http.ResponseWriter, r *http.Request) {
			var message FileRequest
			json.NewDecoder(r.Body).Decode(&message)
			if UidIsValidHash(message.HashValue) {
				go server.DownloadFile(
					state,
					message.Peer,
					UidToHash(message.HashValue),
					message.Filename)
			}
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

	r.HandleFunc("/routingtable",
		func(w http.ResponseWriter, _ *http.Request) {
			json.NewEncoder(w).Encode(state.GetRoutingTableNames())
		}).Methods("GET")

	r.HandleFunc("/message",
		func(w http.ResponseWriter, _ *http.Request) {
			websrv.messages_lock.Lock()
			defer websrv.messages_lock.Unlock()
			json.NewEncoder(w).Encode(websrv.messages)
			websrv.messages = []Message{}
		}).Methods("GET")

	r.HandleFunc("/searchresults",
		func(w http.ResponseWriter, _ *http.Request) {
			websrv.searchresults_lock.Lock()
			defer websrv.searchresults_lock.Unlock()
			json.NewEncoder(w).Encode(websrv.searchresults)
			websrv.searchresults = []WebSearchResult{}
		}).Methods("GET")

	r.HandleFunc("/private",
		func(w http.ResponseWriter, _ *http.Request) {
			websrv.private_lock.Lock()
			defer websrv.private_lock.Unlock()
			json.NewEncoder(w).Encode(websrv.private)
			websrv.private = []PrivateMessage{}
		}).Methods("GET")

	/* we also serve a bunch of static files */
	r.PathPrefix("/").Handler(
		http.StripPrefix("/", http.FileServer(http.Dir("./gui/dist"))))

	return websrv
}

func (srv *WebServer) Start() error {
	return srv.server.ListenAndServe()
}
