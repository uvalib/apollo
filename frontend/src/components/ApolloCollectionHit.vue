<template>
   <div class="collection">
      <div><a :href="hit.collection_url">{{hit.collection_pid}}</a> - {{hit.collection_title}}</div>
      <ul class="hits">
         <li v-for="hit in hit.hits" :key="hit.pid" class="hit">
            <ApolloSearchHit :hit="hit"/>
         </li>
      </ul>
   </div>
</template>

<script>
  import ApolloSearchHit from '@/components/ApolloSearchHit'      
  export default {
      name: "ApolloCollectionHit",
      components: {
         ApolloSearchHit
      },
      props: {
         hit: Object,
      }
   }
</script>

<style scoped>
div.collection {
   color: #34495e;
}
a {
   color: cornflowerblue;
   text-decoration: none;
   font-weight: bold;
}
b {
   margin-right: 10px;
}
a:hover {
   text-decoration: underline;
}
ul {
   list-style-type: none;
   margin-top: 5px;
}
</style>
