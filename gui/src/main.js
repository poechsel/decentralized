import Vue from 'vue'
import App from './App.vue'
import Buefy from 'buefy'

import { library } from '@fortawesome/fontawesome-svg-core'
import { faSync, faPowerOff, faPlus } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faSync)
library.add(faPowerOff)
library.add(faPlus)

Vue.component('font-awesome-icon', FontAwesomeIcon)


import 'buefy/dist/buefy.css';

Vue.config.productionTip = false
Vue.use(Buefy)

new Vue({
  render: h => h(App)
}).$mount('#app')
