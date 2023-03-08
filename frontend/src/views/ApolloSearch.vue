<template>
   <div class="search">
      <PageHeader main="Search" sub="Search Apollo collections" />
      <div class="content">
         <LoadingSpinner v-if="searchStore.searching" message="Searching collections..." />
         <template v-else>
            <ApolloError v-if="searchStore.errorMsg" :message="searchStore.errorMsg" />
            <div class="results-container" v-else>
               <div class="overview">
                  {{searchStore.searchResults.hits}} hits in {{searchStore.searchResults.results.length}} collection(s) on "{{searchStore.searchQuery}}" in {{searchStore.searchResults.response_time_ms}}ms
               </div>
               <h4 v-if="!searchStore.searchResults.hits">
                  No Results Found
               </h4>
               <div v-else class="results">
                  <div v-for="hit in searchResults.results" :key="hit.pid" class="hit">
                     <ApolloCollectionHit :hit="hit"/>
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
import ApolloCollectionHit from '@/components/ApolloCollectionHit.vue'
import { useSearchStore } from '@/stores/search'

const searchStore = useSearchStore()
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
