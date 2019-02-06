<template>
   <div class="search">
      <PageHeader main="Search" sub="Search Apollo collections" />
      <div class="content">
         <LoadingSpinner v-if="searching" message="Searching collections..." />
         <template v-else>
            <ApolloError v-if="errorMsg" :message="errorMsg" />
            <div class="results-container" v-else>
               <div class="overview">{{this.searchResults.hits}} hits in {{this.searchResults.results.length}} collection(s) on "{{this.searchQuery}}" in {{this.searchResults.response_time_ms}}ms</div>
               <h4 v-if="!searchResults.hits">
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

<script>
   import PageHeader from '@/components/PageHeader'
   import LoadingSpinner from '@/components/LoadingSpinner'
   import ApolloError from '@/views/ApolloError'
   import ApolloCollectionHit from '@/components/ApolloCollectionHit'
   import axios from 'axios'
   import { mapGetters } from 'vuex'
   import { mapMutations } from 'vuex'

   export default {
      name: "ApolloSearch",
      props: {
         passedQuery: String,
      },
      components: {
         PageHeader,
         LoadingSpinner,
         ApolloError,
         ApolloCollectionHit
      },
      computed: { 
         ...mapGetters([
            'searchQuery',
         ])
      },
      data: function () {
         return {
            searching: true,
            errorMsg: null,
            searchResults: null
         }
      },

      mounted: function () {
         this.search()
      },

      // Watch for changes in URL to detect new search term and redo the search
      watch: {
         $route(/*to, from*/) {
            this.search()
         },
      },

      methods: {
         ...mapMutations([
            'setSearchQuery'
         ]),
         search: function () {
            // show searching box and do search...
            this.searching = true
            if (this.passedQuery && this.passedQuery.length > 0) {
               this.setSearchQuery(this.passedQuery)
            }
            axios.get("/api/search?q=" + this.searchQuery).then((response) => {
               this.searchResults = response.data
            }).catch((error) => {
               if (error.response) {
                  this.errorMsg = error.response.data
               } else {
                  this.errorMsg = error
               }
            }).finally(() => {
               this.searching = false
            })
         }
      }
   }
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
