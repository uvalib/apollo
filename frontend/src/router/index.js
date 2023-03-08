import { createRouter, createWebHistory } from 'vue-router'
import CollectionList from '@/views/CollectionList.vue'
import CollectionDetails from '@/views/CollectionDetails.vue'
import ApolloError from '@/views/ApolloError.vue'
import ApolloSearch from '@/views/ApolloSearch.vue'

const router = createRouter({
   history: createWebHistory(import.meta.env.BASE_URL),
   routes: [
      {
         path: '/',
         name: 'home',
         component: CollectionList,
      },
      {
         path: '/collections',
         name: 'collections',
         component: CollectionList,
      },
      {
         name: 'collectiondetail',
         path: '/collections/:id',
         component: CollectionDetails,
      },
      {
         path: '/search',
         component: ApolloSearch,
      },
      {
         path: "/unauthorized",
         name: "unauthorized",
         component: ApolloError,
         props: { message: "You are not authorized to access this site" },
      },
      {
         path: '/:pathMatch(.*)*',
         name: "notfound",
         component: ApolloError,
         props: { message: "The page you requested cannot be found" },
      }
   ]
})

export default router
