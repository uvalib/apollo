<template>
   <span class="search">
      <input type="text" id="search" :placeholder="searchPrompt" @keyup.enter="doSearch"/>
   </span>
</template>

<script>
   export default {
      name: 'ApolloSearch',
      props: {
         global: Boolean
      },
      computed: {
         searchPrompt: function() {
            if (this.global) {
               return "Search All Collections..."
            }
            return "Search Collection..."
         }
      },
      methods: {
         doSearch: function() {
            let val = $("#search").val().trim() 
            if (val.length == 0) return
            alert(val)
         }
      }
   }
</script>

<style scoped>
   span.search {
      position: absolute;
      right: 0;
      bottom: 10px;
      width: 30%;
   }
   #search {
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