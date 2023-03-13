<template>
   <div class="search">
      <PageHeader main="Search" sub="Search Apollo collections" />
      <div class="content">
         <LoadingSpinner v-if="searchStore.searching" message="Searching collections..." />
         <template v-else>
            <ApolloError v-if="searchStore.errorMsg" :message="searchStore.errorMsg" />
            <div class="results-container" v-else>
               <div class="overview">
                  {{searchStore.searchResults.total}} hits in {{searchStore.searchResults.collections.length}} collection(s) on "{{searchStore.query}}" in {{searchStore.searchResults.response_time_ms}}ms
               </div>
               <h4 v-if="searchStore.searchResults.total == 0">
                  No Results Found
               </h4>
               <div v-else class="results">
                  <div v-for="coll in searchStore.searchResults.collections" :key="coll.collection_pid" class="hit">
                     <SearchHit :collection="coll"/>
                  </div>
               </div>
            </div>
         </template>
      </div>
   </div>
</template>

<script setup>
import PageHeader from '@/components/PageHeader.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import ApolloError from '@/views/ApolloError.vue'
import SearchHit from '@/components/SearchHit.vue'
import { useSearchStore } from '@/stores/search'
import { onBeforeMount} from 'vue'

const searchStore = useSearchStore()

onBeforeMount( () => {
   searchStore.search()
})
</script>

<style scoped>
   div.search {
      background: white;
      padding: 20px;
   }

   .results-container {
      width: 85%;
      margin: 15px auto;
   }

   .results-container .overview {
      color: #34495e;
      padding-bottom: 5px;
      font-size: 0.9em;
   }

   .results {
      margin-top: 5px;
      padding: 0;
   }
</style>
