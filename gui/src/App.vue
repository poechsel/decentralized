<template>
<div id="app">
  <div id="container">
    <!--
        <nav  class="navbar is-transparent" role="navigation" aria-label="main navigation" style="margin-bottom: 40px">
          <div class="navbar-brand">
            <span class="navbar-item"> GUI </span>
          </div>
          <div id="navbarBasicExample" class="navbar-menu">
            <div class="navbar-start">
              
              <div class="navbar-item">
                <a class="button" v-on:click="refresh" size="sm"> <font-awesome-icon icon="sync"/> </a>
              </div>
              <span class="navbar-item">Last refresh at {{time_last_update.toTimeString()}}</span>
            </div>
          </div>
        </nav>
        -->
    
    <div class="tile is-ancestor">
      
      
      
      
      <div class="tile is-vertical is-parent">
        
        <div class="tile is-child card"  style="flex-grow: 0">
          <header class="card-header">
            <p class="card-header-title">
              File
            </p>
          </header>
          <div class="card-content">
            <b-field class="file">
              <b-upload v-on:input="upload_file" style="width: 100%">
                <a class="button is-fullwidth">
                  <b-icon icon="upload"></b-icon> <span>Upload</span>
                </a>
              </b-upload>
            </b-field>
            <div class="field">
              <a class="button is-fullwidth" @click="isComponentModalActive = true">
                <b-icon icon="download"></b-icon> <span>Download</span>
              </a>
            </div>
            <b-modal :active.sync="isComponentModalActive" has-modal-card>
              <file-download :peers="routing_table"></file-download>
            </b-modal>
            
          </div>
        </div>
        
        <div class="tile is-child card"  style="flex-grow: 0">
          <header class="card-header">
            <p class="card-header-title">
              File Search
            </p>
          </header>
          <div class="card-content">
            <div style="overflow:scroll; margin-bottom: 20px; height: 250px">
              <ul >
                <li style=" border-bottom: 1px solid #EEEEEE"
                    v-for="(search, index) in searches" :key="index">
                  <div class="level" style="margin: 0">
                    <div class="level-left">
                      {{search.Keywords}}
                    </div>
                    <div class="level-right has-text-weight-bold">
                      {{search.FileName}}
                    </div>
                  </div>
                  <a @click="() => downloadFromSearch(search.MetaHash)">
                    <span class="icon is-small mdi mdi-download"></span> <span>{{search.MetaHash}}</span>
                  </a>
                  
                </li>
              </ul>

            <b-modal :active.sync="searchDownloadModal" has-modal-card>
              <file-download-search :hashvalue="selected_metahash_search"></file-download-search>
            </b-modal>
            </div>
            
            <div class="field has-addons">
              <div class="control" style="width: 100%">
                <input class="input" type="text" placeholder="foo,bar"
                       v-model="search_keywords">
              </div>
              <div class="control">
                <a class="button is-primary" v-on:click="new_search">
                  Search
                </a>
              </div>
            </div>

          </div>
        </div>
        
      </div>


      <div class="tile is-vertical is-7 is-parent">
        
        <div class="tile is-child card" >
          <header class="card-header">
            <div class="level">
              <div class="level-left">
                <div style="width: 20px"></div>
                <b-dropdown>
                  <button class="button" slot="trigger">
                    <i class="fas fa-plus"></i>
                    <b-icon icon="menu-down"></b-icon>
                  </button>
                  
                  <b-dropdown-item
                    v-for="peer in routing_table"
                    v-if="!opened_private_channels.includes(peer)"
                    :key="peer"
                    @click="open_private_channel(peer)">
                    {{peer}}
                  </b-dropdown-item>
                </b-dropdown>
                <p class="card-header-title level-item">
                  Messages
                </p>
              </div>
            </div>
          </header>
          <div class="card-content">
            <b-tabs type="is-toggle" :animated="false" @input="(index) => opened_channel = index" v-model="opened_channel">
              <b-tab-item icon="users" iconPack="fa" label="General">
                <messages :messages=messages
                          @submit-message='(content) => {add_message(content)}'>
                </messages>
              </b-tab-item>
              
              <b-tab-item
                icon="user" iconPack="fa"
                v-for="peer in opened_private_channels"
                :label="peer" :key="peer">
                <messages :messages="private_channels[peer]"
                          @submit-message='(content) => {add_private_message(peer, content)}'>
                  >
                </messages>
              </b-tab-item>
              
            </b-tabs>
          </div>
        </div>
        
      </div>
      
      <div class="tile is-vertical is-parent">
        
        <div class="tile is-child card"  style="flex-grow: 0">
          <div class="navbar-item">
            <a class="button" v-on:click="refresh" size="sm"> <i class="fas fa-sync"/> </a>
            <span class="navbar-item">Last refresh at {{time_last_update.toLocaleTimeString()}}</span>
          </div>
        </div>
        
        <div class="tile is-child card"  style="flex-grow: 0">
          <div class="card-content">
            Connected to <strong> {{server.name}} </strong> 
            at address <strong> {{server.address}} </strong>
          </div>
        </div>
        
        <div class="card is-child tile" style="flex-grow: 0">
          <header class="card-header">
            <p class="card-header-title">
              Peers
            </p>
          </header>
          <div class="card-content">
            <b-taglist>
              <b-tag v-for="peer in peers" :key="peer" >{{peer}}</b-tag>
            </b-taglist>
            <div class="field has-addons">
              <div class="control" style="width: 100%">
                <input class="input" type="text" placeholder="Address"
                       v-model="new_peer_address">
              </div>
              <div class="control">
                <a class="button is-primary" v-on:click="add_peer">
                  Add peer
                </a>
              </div>
            </div>
          </div>
        </div>
        
        
      </div>
    </div>
  </div>
</div>
</template>

<script>
import Messages from './components/Messages.vue'
import FileDownload from './components/FileDownload.vue'
import FileDownloadSearch from './components/FileDownloadSearch.vue'

var request = require('request')
//var x = {'Origin': "foo", 'ID': "4", 'Text': "I am a text"}
//var foo = [x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x]

/* eslint-disable */
export default {
    name: 'app',
    components: {
        Messages,
        FileDownload,
        FileDownloadSearch
    },
    data () {
        return {
            server:{address: "Unknown", name:"Unknown"},
            peers_dns: {},
            peers: [],
            routing_table: ["d", "a", "b", "c"],
            new_peer_address: "",
            new_message: "",
            messages: [],
            time_last_update: new Date(Date.now()),
            opened_private_channels: [],
            private_channels: {},
            opened_channel: 0,
            search_keywords: 0,
            selected_metahash_search: "",
            searches: [
            ],
        }
    },
    methods: {
        open_private_channel: function(peer) {
            if (this.opened_private_channels.includes(peer)) {
                return
            }
            this.opened_private_channels.push(peer)
            this.$nextTick(function() {
                this.opened_channel = this.opened_private_channels.length
            })
            if (!(peer in this.private_channels))
                this.private_channels[peer] = []
        },
        
        load_identity: function() {
            request('http://127.0.0.1:8080/id', (error, response, body) => {
                let r = JSON.parse(body) 
                this.server.name = r["Name"]
                this.server.address = r["Address"]
            })
        },
        
        load_peers: function() {
            request('http://127.0.0.1:8080/node', (error, response, body) => {
                let r = JSON.parse(body) 
                this.peers = []
                for (var p of r) {
                    this.peers.push(p)
                }
            })
        },
        get_new_messages: function() {
            request('http://127.0.0.1:8080/message', (error, response, body) => {
                let r = JSON.parse(body) 
                for (var p of r) {
                    this.messages.push(p.Rumor)
                }
            })
        },
        get_search_results: function() {
            request('http://127.0.0.1:8080/searchresults', (error, response, body) => {
                let r = JSON.parse(body) 
                for (var p of r) {
                    this.searches.push(p)
                }
            })
        },
        
        get_new_private_messages: function() {
            request('http://127.0.0.1:8080/private', (error, response, body) => {
                let r = JSON.parse(body) 
                for (var p of r) {
                    let bucket = p.Origin
                    if (bucket == this.server.name) {
                        bucket = p.Destination
                    }
                    if (bucket in this.private_channels) {
                        this.private_channels[bucket].push(p)
                    } else {
                        this.private_channels[bucket] = [p]
                    }
                }
            })
        },
        
        upload_file: function(file) {
            request.post({
                headers: {'content-type' : 'application/json'},
                url:     'http://127.0.0.1:8080/upload',
                body:    JSON.stringify(file.name)
            }, function(error, response, body){
            });
        },
        
        
        get_routing_table: function() {
            request('http://127.0.0.1:8080/routingtable', (error, response, body) => {
                let r = JSON.parse(body) 
                // adding elements one by one to make sure vuejs has seen the
                // update on the array
                this.routing_table = []
                for (var p of r) {
                    this.routing_table.push(p)
                }
                this.routing_table.sort()
            })
        },
        
        downloadFromSearch: function(metahash) {
            this.searchDownloadModal = true
            this.selected_metahash_search = metahash
        },
        
        add_peer: function(event) {
            if (this.new_peer_address != "") {
                request.post({
                    headers: {'content-type' : 'application/json'},
                    url:     'http://127.0.0.1:8080/node',
                    body:    JSON.stringify(this.new_peer_address)
                }, function(error, response, body){
                });
                this.new_peer_address = ""
            }
        },

        new_search: function(event) {
            if (this.search_keywords != "") {
                request.post({
                    headers: {'content-type' : 'application/json'},
                    url:     'http://127.0.0.1:8080/search',
                    body:    JSON.stringify(this.search_keywords)
                }, function(error, response, body){
                });
                this.new_peer_address = ""
            }
        },
        
        add_message: function(content) {
            request.post({
                headers: {'content-type' : 'application/json'},
                url:     'http://127.0.0.1:8080/message',
                body:    JSON.stringify(content)
            }, function(error, response, body){
            });
        },
        
        add_private_message: function(peer, content) {
            request.post({
                headers: {'content-type' : 'application/json'},
                url:     'http://127.0.0.1:8080/private',
                body:    JSON.stringify({'Content':content, 'To': peer})
            }, function(error, response, body){
            });
        },
        
        refresh: function() {
            this.load_peers()
            this.get_new_messages()
            this.get_new_private_messages()
            this.get_routing_table()
            this.get_search_results()
            this.time_last_update = new Date(Date.now())
        }
    },
    computed: {
        peers_map: function() {
            let pm = {}
            for (var p of this.peers) {
                if (p in this.peers_dns) {
                    pm[p] = this.peers_dns[p]
                } else {
                    pm[p] = "Unknown"
                }
            }
            return pm
        },
    },
    mounted: function() {
        this.load_identity()
        this.refresh();
        
        setInterval(() => {
            this.refresh();
        }, 1000);
    },
}

</script>

<style>
#app {
    height: 100%;
    padding: 150px;
    padding-top: 100px;
    padding-bottom: 100px;
}

#container {
    font-family: 'Avenir', Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    text-align: center;
    margin: auto;
    height: 100%;
    box-sizing: border-box; 
}

.scrolly {
    overflow-y: auto;
}

.ovbody {
    overflow-y: auto;
    }

</style>
