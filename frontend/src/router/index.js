import Vue from 'vue'
import Router from 'vue-router'
import CollectionList from '@/components/CollectionList'
import CollectionDetails from '@/components/CollectionDetails'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: CollectionList
    },
    {
      name: 'collections',
      path: '/collections/:id',
      component: CollectionDetails,
      props: true
    }
  ]
})
