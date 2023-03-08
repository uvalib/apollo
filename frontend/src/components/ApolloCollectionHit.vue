<template>
   <div class="collection">
      <div><a :href="props.hit.collection_url">{{ props.hit.collection_pid }}</a> - {{ props.hit.collection_title }}</div>
      <ul class="hits">
         <li v-for="hit in props.hit.hits" :key="props.hit.pid" class="hit">
            <ApolloSearchHit :hit="hit" />
         </li>
      </ul>
   </div>
</template>

<script setup>
import ApolloSearchHit from '@/components/ApolloSearchHit.vue'
const props = defineProps({
   hit: {
      type: Object,
      required: true
   }
})
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
}</style>
