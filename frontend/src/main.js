// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import LoadingSpinner from './components/LoadingSpinner'
import ApolloError from './components/ApolloError'
import router from './router'

require('./assets/css/shared.css');

Vue.config.productionTip = false;

// global register some components thar all pages use
Vue.component("loading-spinner", LoadingSpinner)
Vue.component("apollo-error", ApolloError)

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  components: { App },
  template: '<App/>'
});
