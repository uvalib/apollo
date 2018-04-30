import Vue from 'vue'
import Router from 'vue-router'
import CollectionList from '@/components/CollectionList'
import CollectionDetail from '@/components/CollectionDetail'

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
      component: CollectionDetail,
      props: true
    }
  ]
})
