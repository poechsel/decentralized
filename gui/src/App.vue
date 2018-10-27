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
      <div class="tile is-vertical is-8 is-parent">
        
        <div class="tile is-child card" >
          <header class="card-header">
            <p class="card-header-title">
              Messages
            </p>
          </header>
          <div class="card-content">
            <div style="overflow:hidden; margin-bottom: 20px; height: 500px">
              <ul >
                <li style="padding: 10px 40px 10px 40px; border-bottom: 1px solid #EEEEEE"
                    v-for="message in messages" :key="message.Address + message.Rumor.ID">
                  <div class="level">
                    <div class="level-left">
                      <b-taglist attached>
                        <b-tag type="is-info">{{message.Rumor.Origin}}</b-tag>
                        <b-tag type="is-light">{{message.Rumor.ID}}</b-tag>
                      </b-taglist>
                    </div>
                    <div class="level-right">
                      <span class="d-flex justify-content-end">
                        {{message.Rumor.Text}}
                      </span>
                    </div>
                  </div>
                </li>
              </ul>
            </div>
            
            <div class="field has-addons">
              <div class="control" style="width:100%">
                <input class="input" type="text" placeholder="Enter message"
                       v-model="new_message">
              </div>
              <div class="control">
                <a class="button is-info" v-on:click="add_message">
                  Send
                </a>
              </div>
            </div>
          </div>
        </div>
        
      </div>
      
      <div class="tile is-vertical is-parent">

        <div class="tile is-child card"  style="flex-grow: 0">
          <div class="navbar-item">
            <a class="button" v-on:click="refresh" size="sm"> <font-awesome-icon icon="sync"/> </a>
          <span class="navbar-item">Last refresh at {{time_last_update.toTimeString()}}</span>
          </div>
        </div>

        <div class="tile is-child card"  style="flex-grow: 0">
          <div class="card-content">
            Connected to <strong> {{server.name}} </strong> 
            at address <strong> {{server.address}} </strong>
          </div>
        </div>
        
        <div class="card is-child tile">
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
                <a class="button is-info" v-on:click="add_peer">
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
//import HelloWorld from './components/HelloWorld.vue'

var request = require('request')
var x = {'Address': "", 'Rumor':{'Origin': "foo", 'ID': "4", 'Text': "I am a text"}}
var foo = [x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x]
 
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
            peers: ["arzer", "ztoetih"],
            new_peer_address: "",
            new_message: "",
            messages: foo,
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
