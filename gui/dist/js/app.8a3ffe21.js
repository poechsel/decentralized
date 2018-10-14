(function(e){function t(t){for(var n,o,i=t[0],u=t[1],d=t[2],c=0,p=[];c<i.length;c++)o=i[c],s[o]&&p.push(s[o][0]),s[o]=0;for(n in u)Object.prototype.hasOwnProperty.call(u,n)&&(e[n]=u[n]);l&&l(t);while(p.length)p.shift()();return a.push.apply(a,d||[]),r()}function r(){for(var e,t=0;t<a.length;t++){for(var r=a[t],n=!0,i=1;i<r.length;i++){var u=r[i];0!==s[u]&&(n=!1)}n&&(a.splice(t--,1),e=o(o.s=r[0]))}return e}var n={},s={app:0},a=[];function o(t){if(n[t])return n[t].exports;var r=n[t]={i:t,l:!1,exports:{}};return e[t].call(r.exports,r,r.exports,o),r.l=!0,r.exports}o.m=e,o.c=n,o.d=function(e,t,r){o.o(e,t)||Object.defineProperty(e,t,{enumerable:!0,get:r})},o.r=function(e){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},o.t=function(e,t){if(1&t&&(e=o(e)),8&t)return e;if(4&t&&"object"===typeof e&&e&&e.__esModule)return e;var r=Object.create(null);if(o.r(r),Object.defineProperty(r,"default",{enumerable:!0,value:e}),2&t&&"string"!=typeof e)for(var n in e)o.d(r,n,function(t){return e[t]}.bind(null,n));return r},o.n=function(e){var t=e&&e.__esModule?function(){return e["default"]}:function(){return e};return o.d(t,"a",t),t},o.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},o.p="/";var i=window["webpackJsonp"]=window["webpackJsonp"]||[],u=i.push.bind(i);i.push=t,i=i.slice();for(var d=0;d<i.length;d++)t(i[d]);var l=u;a.push([0,"chunk-vendors"]),r()})({0:function(e,t,r){e.exports=r("56d7")},"034f":function(e,t,r){"use strict";var n=r("c21b"),s=r.n(n);s.a},1:function(e,t){},2:function(e,t){},3:function(e,t){},4:function(e,t){},5:function(e,t){},"56d7":function(e,t,r){"use strict";r.r(t);r("cadf"),r("551c"),r("097d");var n=r("2b0e"),s=function(){var e=this,t=e.$createElement,r=e._self._c||t;return r("div",{attrs:{id:"app"}},[r("div",{attrs:{id:"container"}},[r("b-navbar",{staticClass:"mb-3",attrs:{variant:"faded",type:"light"}},[r("b-navbar-brand",{staticClass:"mb-0",attrs:{tag:"h1"}},[e._v("GUI")]),r("b-button",{attrs:{size:"sm"},on:{click:e.refresh}},[r("font-awesome-icon",{attrs:{icon:"sync"}})],1),r("b-nav-text",{staticStyle:{"padding-left":"10px"}},[e._v("Last refresh at "+e._s(e.time_last_update.toTimeString()))]),r("b-navbar-nav",{staticClass:"ml-auto"})],1),r("b-container",{staticClass:"mb-3"},[r("b-row",[r("b-col",{attrs:{cols:"8"}},[r("b-card",{staticStyle:{height:"600px"},attrs:{header:"Messages","body-class":"ovbody"}},[r("b-list-group",{attrs:{flush:""}},e._l(e.messages,function(t){return r("b-list-group-item",{key:t.Address+t.Rumor.ID,attrs:{clas:"d-flex flex-row"}},[r("div",{staticClass:"d-flex justify-content-start"},[r("b-badge",{attrs:{variant:"primary",pill:""}},[e._v(e._s(t.Rumor.ID))]),r("strong",[e._v(e._s(t.Rumor.Origin))])],1),r("div",{staticClass:"d-flex justify-content-end"},[e._v("\n                  "+e._s(t.Rumor.Text)+"\n                ")])])}))],1),r("b-form",{on:{submit:function(e){e.preventDefault()}}},[r("b-input-group",[r("b-form-input",{attrs:{id:"message",required:"",placeholder:"Enter message"},model:{value:e.new_message,callback:function(t){e.new_message=t},expression:"new_message"}}),r("b-button",{attrs:{type:"button",variant:"primary"},on:{click:e.add_message}},[e._v("Send")])],1)],1)],1),r("b-col",[r("b-card",{staticClass:"mb-2",attrs:{header:"identity"}},[r("strong",[e._v("Address: ")]),e._v(" "+e._s(e.server.address)+" "),r("br"),r("strong",[e._v("Name: ")]),e._v(" "+e._s(e.server.name)+"\n        ")]),r("b-card",{staticStyle:{height:"400px"},attrs:{header:"Peers","body-class":"ovbody"}},[r("b-list-group",{attrs:{flush:""}},e._l(e.peers_map,function(t,n){return r("b-list-group-item",{key:t+n},[r("strong",[e._v(e._s(t))]),e._v("\n              "+e._s(n)+"\n            ")])}))],1),r("b-form",{on:{submit:function(e){e.preventDefault()}}},[r("b-input-group",[r("b-form-input",{attrs:{id:"peer_address",required:"",placeholder:"Enter peer address"},model:{value:e.new_peer_address,callback:function(t){e.new_peer_address=t},expression:"new_peer_address"}}),r("b-button",{attrs:{type:"button",variant:"primary"},on:{click:e.add_peer}},[e._v("Add Peer")])],1)],1)],1)],1)],1)],1)])},a=[],o=(r("ac4d"),r("8a81"),r("ac6a"),r("30dc")),i={name:"app",components:{},data:function(){return{server:{address:"Unknown",name:"Unknown"},peers_dns:{},peers:[],new_peer_address:"",new_message:"",messages:[],time_last_update:new Date(Date.now())}},methods:{load_identity:function(){var e=this;o("http://127.0.0.1:8080/id",function(t,r,n){var s=JSON.parse(n);e.server.name=s["Name"],e.server.address=s["Address"]})},load_peers:function(){var e=this;o("http://127.0.0.1:8080/node",function(t,r,n){var s=JSON.parse(n);e.peers=[];var a=!0,o=!1,i=void 0;try{for(var u,d=s[Symbol.iterator]();!(a=(u=d.next()).done);a=!0){var l=u.value;e.peers.push(l)}}catch(e){o=!0,i=e}finally{try{a||null==d.return||d.return()}finally{if(o)throw i}}})},get_new_messages:function(){var e=this;o("http://127.0.0.1:8080/message",function(t,r,n){var s=JSON.parse(n),a=!0,o=!1,i=void 0;try{for(var u,d=s[Symbol.iterator]();!(a=(u=d.next()).done);a=!0){var l=u.value;e.messages.push(l)}}catch(e){o=!0,i=e}finally{try{a||null==d.return||d.return()}finally{if(o)throw i}}})},add_peer:function(e){""!=this.new_peer_address&&(o.post({headers:{"content-type":"application/json"},url:"http://127.0.0.1:8080/node",body:JSON.stringify(this.new_peer_address)},function(e,t,r){}),this.new_peer_address="")},add_message:function(e){""!=this.new_message&&(o.post({headers:{"content-type":"application/json"},url:"http://127.0.0.1:8080/message",body:JSON.stringify(this.new_message)},function(e,t,r){}),this.new_message="")},refresh:function(){this.load_peers(),this.get_new_messages(),this.time_last_update=new Date(Date.now())}},computed:{peers_map:function(){var e={},t=!0,r=!1,n=void 0;try{for(var s,a=this.peers[Symbol.iterator]();!(t=(s=a.next()).done);t=!0){var o=s.value;o in this.peers_dns?e[o]=this.peers_dns[o]:e[o]="Unknown"}}catch(e){r=!0,n=e}finally{try{t||null==a.return||a.return()}finally{if(r)throw n}}return e}},mounted:function(){var e=this;this.load_identity(),this.refresh(),setInterval(function(){e.refresh()},1e3)}},u=i,d=(r("034f"),r("2877")),l=Object(d["a"])(u,s,a,!1,null,null,null);l.options.__file="App.vue";var c=l.exports,p=r("9f7b"),f=r("ecee"),v=r("c074"),b=r("7a55");r("f9e3"),r("2dd8");f["library"].add(v["c"]),f["library"].add(v["b"]),f["library"].add(v["a"]),n["a"].component("font-awesome-icon",b["FontAwesomeIcon"]),n["a"].config.productionTip=!1,n["a"].use(p["a"]),new n["a"]({render:function(e){return e(c)}}).$mount("#app")},c21b:function(e,t,r){}});
//# sourceMappingURL=app.8a3ffe21.js.map