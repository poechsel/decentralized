<template>
<div id="app">
  <div id="container">
  <b-navbar variant="faded" type="light" class="mb-3">
    <b-navbar-brand tag="h1" class="mb-0">GUI</b-navbar-brand>
    <b-button size="sm"> <font-awesome-icon icon="sync"/> </b-button>
    <b-nav-text style="padding-left: 10px;">Last refresh x second ago</b-nav-text>
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
            <b-list-group-item v-for="item in items" :key="item.peer + item.id">
              <strong>{{item.peer}}</strong>
              <b-badge variant="primary" pill>{{item.id}}</b-badge>
              {{item.message}}
            </b-list-group-item>
          </b-list-group>
        </b-card>
        <b-form>
          <b-input-group>
            <b-form-input id="exampleInput1"
                          type="email"
                          v-model="form.email"
                          required
                          placeholder="Enter email">
            </b-form-input>
            <b-button type="submit" variant="primary">Submit</b-button>
          </b-input-group>
        </b-form>
      </b-col>
      <b-col>
        
        <b-card header="identity" class="mb-2">
          <strong>Address: </strong> x.y.z <br>
          <strong>Name: </strong> ezt
        </b-card>

        <b-card header="Peers" style="height: 400px;" body-class="ovbody">
          <b-list-group flush>
            <b-list-group-item v-for="item in items" :key="item.peer + item.id">
              <strong>{{item.peer}}</strong>
              <b-badge variant="primary" pill>{{item.id}}</b-badge>
              {{item.message}}
            </b-list-group-item>
          </b-list-group>
        </b-card>
        <b-form>
          <b-input-group>
            <b-form-input id="exampleInput1"
                          type="email"
                          v-model="form.email"
                          required
                          placeholder="Enter email">
            </b-form-input>
            <b-button type="submit" variant="primary">Submit</b-button>
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
const items = [
    { peer: true, id: 40, message: 'Macdonald' },
    { peer: false, id: 21, message: 'Shaw' },
    { peer: false, id: 89, message: 'Wilson' },
    { peer: true, id: 38, message: 'Carney' },
    { peer: true, id: 40, message: 'Macdonald' },
    { peer: false, id: 21, message: 'Shaw' },
    { peer: false, id: 89, message: 'Wilson' },
    { peer: true, id: 38, message: 'Carney' },
    { peer: true, id: 40, message: 'Macdonald' },
    { peer: false, id: 21, message: 'Shaw' },
    { peer: false, id: 89, message: 'Wilson' },
    { peer: true, id: 38, message: 'Carney' },
    { peer: true, id: 40, message: 'Macdonald' },
    { peer: false, id: 21, message: 'Shaw' },
    { peer: false, id: 89, message: 'Wilson' },
    { peer: true, id: 38, message: 'Carney' },
    { peer: false, id: 89, message: 'Wilson' },
    { peer: true, id: 38, message: 'Carney' },
    { peer: true, id: 40, message: 'Macdonald' },
    { peer: false, id: 21, message: 'Shaw' },
    { peer: false, id: 89, message: 'Wilson' },
    { peer: true, id: 38, message: 'Carney' },
]

export default {
    name: 'app',
    components: {
        //    HelloWorld
    },
    data () {
        return {
            form: {email: ""},
            items: items,
        }
    }
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
