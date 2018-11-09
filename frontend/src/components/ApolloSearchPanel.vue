<template>
   <span class="search-panel">
      <input type="text" id="search-term" :value="query" placeholder="Search all collections..." @keyup.enter="doSearch"/>
   </span>
</template>

<script>
   import router from '../router'
   export default {
      name: 'ApolloSearchPanel',
      props: {
         query: String
      },
      methods: {
         doSearch: function() {
            let val = $("#search-term").val().trim() 
            if (val.length == 0) return
            router.push({ path: '/search', query: { q: val }})
         }
      }
   }
</script>

<style scoped>
   span.search-panel {
      position: absolute;
      right: 0;
      bottom: 10px;
      width: 30%;
   }
   #search-term {
      border-radius: 20px;
      width: 100%;
      box-sizing: border-box;
      border: 1px solid #ccc;
      padding: 4px 12px;
      color: #666;
      font-size: 0.9em;
      outline: none;
   }
</style>