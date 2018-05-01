import Vue from 'vue'
import Router from 'vue-router'
import CollectionList from '@/components/CollectionList'
import CollectionDetails from '@/components/CollectionDetails'
import ApolloError from '@/components/ApolloError'

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
      path: '/collections',
      component: CollectionList
    },
    {
      name: 'collections',
      path: '/collections/:id',
      component: CollectionDetails,
      props: true
    },
    {
      path: "*",
      component: ApolloError,
      props: { message: "The page you requested cannot be found" }
    }
  ]
})
