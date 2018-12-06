<template>
  <div>
    <div style="overflow:scroll; margin-bottom: 20px; height: 500px">
      <ul >
        <li style="padding: 10px 40px 10px 40px; border-bottom: 1px solid #EEEEEE"
            v-for="(message, index) in messages" :key="index">
          <div class="level">
            <div class="level-left">
              <b-taglist attached>
                <b-tag type="is-primary">{{message.Origin}}</b-tag>
                <b-tag type="is-light">{{message.ID}}</b-tag>
              </b-taglist>
            </div>
            <div class="level-right">
              <span class="d-flex justify-content-end">
                {{message.Text}}
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
        <a class="button is-primary" v-on:click="add_message">
          Send
        </a>
      </div>
    </div>
  </div>
</template>

<script>
  //import HelloWorld from './components/HelloWorld.vue'
  /* eslint-disable */
export default {
    components: {
        //    HelloWorld
    },
    props: ['messages'],
    data () {
        return {
            new_message: "",
        }
    },
    methods: {
        add_message: function(event) {
            if (this.new_message != "") {
                this.$emit('submit-message', this.new_message)
                this.new_message = ""
            }
        },
        
        refresh: function() {
            this.load_peers()
            this.get_new_messages()
            this.time_last_update = new Date(Date.now())
        }
        
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
