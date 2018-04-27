import Vue from 'vue'
import Router from 'vue-router'
import CollectionList from '@/components/CollectionList'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: CollectionList
    }
  ]
})
