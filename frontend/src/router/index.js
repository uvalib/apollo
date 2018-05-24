import Vue from 'vue'
import Router from 'vue-router'
import CollectionList from '@/components/CollectionList'
import CollectionDetails from '@/components/CollectionDetails'
import ApolloError from '@/components/ApolloError'

Vue.use(Router)

export default new Router({
  // mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: CollectionList,
      meta: { requiresAuth: true }
    },
    {
      path: '/collections',
      component: CollectionList,
      meta: { requiresAuth: true }
    },
    {
      name: 'collections',
      path: '/collections/:id',
      component: CollectionDetails,
      meta: { requiresAuth: true },
      props: true
    },
    {
      path: "/unauthorized",
      name: "unauthorized",
      component: ApolloError,
      props: { message: "You are not authorized to access this site" },
      meta: { requiresAuth: false }
    },
    {
      path: "*",
      name: "notfound",
      component: ApolloError,
      props: { message: "The page you requested cannot be found" },
      meta: { requiresAuth: false }
    }
  ]
})
