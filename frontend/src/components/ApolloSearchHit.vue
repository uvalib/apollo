<template>
   <div class="hit">
      <div><a :href="props.hit.item_url">{{ props.hit.pid }}</a> - {{ hitTitle }}</div>
      <p v-html="hitSnippet"></p>
   </div>
</template>

<script setup>
import { useSearchStore } from '@/stores/search'

const searchStore = useSearchStore()

const props = defineProps({
   hit: {
      type: Object,
      required: true
   }
})

const hitTitle = (()=>{
   if (props.hit.title) {
      return props.hit.title
   } else {
      return props.hit.match
   }
})
const hitSnippet = (()=>{
   let matchStr = props.hit.match
   let lcMatch = props.hit.match.toLowerCase()
   let p0 = lcMatch.indexOf(this.searchQuery)
   let p1 = p0 + this.searchQuery.length
   let tStyle = "text-transform:capitalize;font-weight:600"
   let hStyle = "color:rgb(229, 114, 0);font-weight:500;"
   let out = `<span style="${tStyle}">${props.hit.match_type}: </span>${matchStr.substring(0, p0)}<b style="${hStyle}">${matchStr.substring(p0, p1)}</b>${matchStr.substring(p1)}`
   return out
})
</script>

<style scoped>
p {
   margin: 3px 0 15px 0;
}

div.hit {
   font-size: 0.85em;
   padding-bottom: 5px;
   color: #666;
}

a {
   color: cornflowerblue;
   text-decoration: none;
   font-weight: bold;
}

a:hover {
   text-decoration: underline;
}</style>
