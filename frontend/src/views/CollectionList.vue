<template>
   <ApolloError v-if="collectionsStore.error" :message="collectionsStore.error" />
   <div v-else class="collections">
      <PageHeader main="Collections" sub="The following are all of the digitized serials managed by Apollo" />
      <LoadingSpinner v-if="collectionsStore.loading" message="Loading collections" />
      <div v-else class="content">
         <table class="collection-list">
            <tr>
               <th></th>
               <th class="right">PID</th>
               <th>Title</th>
            </tr>
            <tr v-for="item in collectionsStore.collections" :key="item.pid">
               <td class="icon">
                  <router-link :to="`/collections/${item.pid}`" @Click="itemClicked(item)" >
                     <img class="detail" src="../assets/detail.png" />
                  </router-link>
               </td>
               <td class="right">{{ item.pid }}</td>
               <td>{{ item.title }}</td>
            </tr>
         </table>
      </div>
   </div>
</template>

<script setup>
import ApolloError from '@/views/ApolloError.vue'
import PageHeader from '@/components/PageHeader.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useCollectionsStore } from '@/stores/collections'
import { onBeforeMount } from 'vue'

const collectionsStore = useCollectionsStore()

onBeforeMount(async () => {
   collectionsStore.getCollections()
})

const itemClicked = ((item) => {
   collectionsStore.collectionSelected(item.pid)
})
</script>

<style scoped>
div.collections {
   background: white;
   padding: 20px;
}

table {
   margin: 15px;
   border-collapse: collapse;
   border-right: 1px solid #ddd;
   border-left: 1px solid #ddd;
}

td.right,
th.right {
   text-align: right;
   padding-right: 10px;
   border-right: 1px solid #ccc;
}

td.icon {
   padding: 0 10px;
}

img.detail {
   vertical-align: middle;
   opacity: 0.6;
   cursor: pointer;
}

img.detail:hover {
   opacity: 1;
}

td {
   cursor: default;
   padding: 10px 10px 8px 10px;
   border-bottom: 1px solid #ddd;
}

th {
   background-color: #f5f5f5;
   padding: 10px 10px 8px 10px;
   text-align: left;
   border-bottom: 1px solid #ccc;
   border-top: 1px solid #ddd;
}</style>
