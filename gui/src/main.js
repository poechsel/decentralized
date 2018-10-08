import Vue from 'vue'
import App from './App.vue'
import BootstrapVue from 'bootstrap-vue'

import { library } from '@fortawesome/fontawesome-svg-core'
import { faSync, faPowerOff, faPlus } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faSync)
library.add(faPowerOff)
library.add(faPlus)

Vue.component('font-awesome-icon', FontAwesomeIcon)


import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'

Vue.config.productionTip = false
Vue.use(BootstrapVue);

new Vue({
  render: h => h(App)
}).$mount('#app')
