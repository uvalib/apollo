import Vue from 'vue'
import Router from 'vue-router'
import Collections from '@/components/Collections'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: Collections
    }
  ]
})
