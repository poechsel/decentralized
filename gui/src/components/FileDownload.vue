<template>
        <div class="modal-card" style="width: auto">
            <header class="modal-card-head">
                <p class="modal-card-title">Download</p>
            </header>
            <section class="modal-card-body">
              <b-field label="From Peer">
                <b-select placeholder="Select a peer" v-model="selectedpeer" expanded required>
                  <option
                    v-for="peer in peers"
                    :value="peer"
                    :key="peer" required>
                    {{ peer }}
                  </option>
                </b-select>
              </b-field>
              
              <b-field label="Filename">
                <b-input
                  v-model="filename"
                  placeholder="Downloaded file name"
                  required>
                </b-input>
              </b-field>
              
              <b-field label="Metahash">
                <b-input
                  v-model="hashvalue"
                  placeholder="Hash of the metafile"
                  maxlength=64
                  required>
                </b-input>
              </b-field>

            </section>
            <footer class="modal-card-foot">
                <button class="button" type="button" @click="$parent.close()">Close</button>
                <a class="button is-primary" @click="upload">Upload</a>
            </footer>
        </div>
    </template>

<script>
var request = require('request')
/* eslint-disable */
export default {
    props: ['email', 'password', 'peers'],
    data () {
        return {
            hashvalue: "",
            filename: "",
            selectedpeer: "",
        }
    },
    methods: {
        upload: function(e) {
            e
            if (this.selectedpeer != "" && this.filename != "" && this.hashvalue != "") {
                let form = {'Peer': this.selectedpeer,
                            'Filename': this.filename,
                            'HashValue': this.hashvalue,
                           }
                request.post({
                    headers: {'content-type' : 'application/json'},
                    url:     'http://127.0.0.1:8080/download',
                    body:    JSON.stringify(form)
                }, function(error, response, body){
                });
                this.$parent.close()
            }
        },
    }
  }
</script>

<style>
</style>
