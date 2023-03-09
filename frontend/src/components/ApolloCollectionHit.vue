<template>
   <div class="collection">
      <div class="link pad-bottom">
         <a :href="props.collection.collection_url">{{ props.collection.collection_pid }}</a>
         <span class="title">: {{ props.collection.collection_title }}</span>
      </div>
      <ul class="hits">
         <li v-for="hit in props.collection.hits" :key="hit.pid" class="hit">
            <div class="hit">
               <div class="link">
                  <a :href="hit.item_url">{{ hit.pid }}</a>
               </div>
               <p class="hit-title">{{ hitTitle(hit) }}</p>
               <p class="hit-title" v-html="hitSnippet(hit)"></p>
            </div>
         </li>
      </ul>
   </div>
</template>

<script setup>
import { useSearchStore } from '@/stores/search'

const searchStore = useSearchStore()

const props = defineProps({
   collection: {
      type: Object,
      required: true
   }
})

const hitTitle = (( hit ) =>{
   if (hit.title) {
      return hit.title
   } else {
      return hit.match
   }
})

const hitSnippet = ((hit)=>{
   let matchStr = hit.match
   let lcMatch = hit.match.toLowerCase()
   let p0 = lcMatch.indexOf(searchStore.query)
   let p1 = p0 + searchStore.query.length
   let tStyle = "text-transform:capitalize;font-weight:600"
   let hStyle = "color:rgb(229, 114, 0);font-weight:500;"
   let out = `<span style="${tStyle}">Matched ${hit.match_type}: </span>${matchStr.substring(0, p0)}<b style="${hStyle}">${matchStr.substring(p0, p1)}</b>${matchStr.substring(p1)}`
   return out
})
</script>

<style scoped lang="scss">
div.collection {
   color: #34495e;
   .link {
      .title {
         display: inline-block;
         margin-left: 5px;
      }
      a {
         color: cornflowerblue;
         text-decoration: none;
         font-weight: bold;
         &:hover {
            text-decoration: underline;
         }
      }
   }
   .pad-bottom {
      margin-bottom: 10px;
   }
   ul {
      list-style-type: none;
      margin-top: 5px;
      .hit-title {
         margin:0 0 5px 20px;
         font-size: 0.9em;
      }
   }
}
</style>
