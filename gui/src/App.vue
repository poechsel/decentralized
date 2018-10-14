<template>
<div id="app">
  <div id="container">
    <b-navbar variant="faded" type="light" class="mb-3">
      <b-navbar-brand tag="h1" class="mb-0">GUI</b-navbar-brand>
      <b-button size="sm"> <font-awesome-icon icon="sync"/> </b-button>
      <b-nav-text style="padding-left: 10px;">Last refresh at {{time_last_update.toTimeString()}}</b-nav-text>
      <b-navbar-nav class="ml-auto">
        <b-button size="sm">  <font-awesome-icon icon="power-off"/> </b-button>
      </b-navbar-nav>
    </b-navbar>
    <!-- HelloWorld msg="Welcome to Your Vue.js App"/-->
    <b-container class="bv-example-row mb-3">
      <b-row>
        <b-col cols="8">
          <b-card header="Messages" style="height: 600px;" body-class="ovbody">
            <b-list-group flush>
              <b-list-group-item clas="d-flex flex-row" v-for="message in messages" :key="message.Address + message.Rumor.ID">
                <div class="d-flex justify-content-start">
                  <b-badge variant="primary" pill>{{message.Rumor.ID}}</b-badge>
                  <strong>{{message.Rumor.Origin}}</strong>
                </div>
                <div class="d-flex justify-content-end">
                  {{message.Rumor.Text}}
                </div>
              </b-list-group-item>
          </b-list-group>
        </b-card>
        <b-form>
          <b-input-group>
            <b-form-input id="message"
                          v-model="new_message"
                          required
                          placeholder="Enter message">
            </b-form-input>
            <b-button type="button" v-on:click="add_message" variant="primary">Send</b-button>
          </b-input-group>
        </b-form>
      </b-col>
      <b-col>
        
        <b-card header="identity" class="mb-2">
          <strong>Address: </strong> {{server.address}} <br>
          <strong>Name: </strong> {{server.name}}
        </b-card>

        <b-card header="Peers" style="height: 400px;" body-class="ovbody">
          <b-list-group flush>
            <b-list-group-item v-for="(peer, name) in peers_map" :key="peer + name">
              <strong>{{peer}}</strong>
              {{name}}
            </b-list-group-item>
          </b-list-group>
        </b-card>
        <b-form>
          <b-input-group>
            <b-form-input id="peer_address"
                          v-model="new_peer_address"
                          required
                          placeholder="Enter peer address">
            </b-form-input>
            <b-button type="button" v-on:click="add_peer" variant="primary">Add Peer</b-button>
          </b-input-group>
        </b-form>
        
      </b-col>
    </b-row>
  </b-container>
</div>
</div>
</template>

<script>
//import HelloWorld from './components/HelloWorld.vue'

var request = require('request')

/* eslint-disable */
export default {
    name: 'app',
    components: {
        //    HelloWorld
    },
    data () {
        return {
            server:{address: "Unknown", name:"Unknown"},
            peers_dns: {},
            peers: [],
            new_peer_address: "",
            new_message: "",
            messages: [],
            time_last_update: new Date(Date.now()),
        }
    },
    methods: {
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
                    this.messages.push(p)
                }
            })
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
        
        add_message: function(event) {
            if (this.new_message != "") {
                request.post({
                    headers: {'content-type' : 'application/json'},
                    url:     'http://127.0.0.1:8080/message',
                    body:    JSON.stringify(this.new_message)
                }, function(error, response, body){
                });
                this.new_message = ""
            }
        },

        refresh: function() {
            this.load_peers()
            this.get_new_messages()
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
        }, 2000);
    },
}

</script>

<style>
#container {
    font-family: 'Avenir', Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    text-align: center;
    color: #2c3e50;
    width: 1000px;
    margin: auto;
    overflow: hidden;
    background-color: white;
    margin-top: 60px;
    margin-bottom: 60px;
    height: 100%;
    
    box-sizing: border-box; 
}


.wrapper_msg {
    overflow-y: scroll;
    height: calc(80% - 100px);
    margin-bottom: 20px;
}

.scrolly {
    overflow-y: auto;
}

.ovbody {
    overflow-y: auto;
    margin: 0 !important;
    padding: 0 !important;
    }

</style>
