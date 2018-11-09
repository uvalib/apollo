<template>
   <div class="search">
      <PageHeader main="Search" sub="Search Apollo collections" :query="query" />
      <div class="content">
         <LoadingSpinner v-if="searching" message="Searching collections..." />
         <template v-else>
            <ApolloError v-if="errorMsg" :message="errorMsg" />
            <div class="results-container" v-else>
               <div class="overview">{{this.searchResults.hits}} hits on "{{this.query}}" in {{this.searchResults.response_time_ms}}ms</div>
               <h4 v-if="!searchResults.hits">
                  No Results Found
               </h4>
               <ul v-else class="results">
                  <li v-for="hit in searchResults.results" :key="hit.pid" class="hit">
                     <ApolloSearchHit :hit="hit"/>
                  </li>
               </ul>
            </div>
         </template>
      </div>
   </div>
</template>

<script>
   import PageHeader from '@/components/PageHeader'
   import LoadingSpinner from '@/components/LoadingSpinner'
   import ApolloError from '@/views/ApolloError'
   import ApolloSearchHit from '@/components/ApolloSearchHit'
   import axios from 'axios'

   export default {
      name: "ApolloSearch",
      components: {
         PageHeader,
         LoadingSpinner,
         ApolloError,
         ApolloSearchHit
      },

      props: {
         query: String,
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
         search: function () {
            // show searching box and do search...
            this.searching = true
            axios.get("/api/search?q=" + this.query).then((response) => {
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
      color: #666;
      padding-bottom: 5px;
      font-size: 0.9em;
   }

   ul.results {
      margin-top: 15px;
      padding: 0;
   }
   li {
      list-style-type:none;
      padding-bottom: 15px;
   }
</style>
