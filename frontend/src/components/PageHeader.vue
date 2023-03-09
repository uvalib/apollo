<template>
   <div class="page-header">
      <h2 class="page-header">{{ main }}</h2>
      <p class="page-header"><span v-html="sub"></span></p>
      <span class="search-panel">
         <input type="text" id="search-term" v-model="searchStore.query" placeholder="Search all collections..." @keyup.enter="doSearch"/>
      </span>
   </div>
</template>

<script setup >
import { useRouter, useRoute } from 'vue-router'
import { useSearchStore } from '@/stores/search'

const searchStore = useSearchStore()
const router = useRouter()
const route = useRoute()

const props = defineProps({
   main: {
      type: String,
      required: true
   },
   sub: {
      type: String,
      default: ""
   },
})

const doSearch = (() =>{
   if ( route.path == "/search") {
      searchStore.search()
   } else {
      router.push("/search")
   }
})
</script>

<style scoped>
div.page-header {
   border-bottom: 1px solid rgb(229, 114, 0);
   margin-bottom: 15px;
   padding-bottom: 5px;
   position: relative;
}

span.user {
   position: absolute;
   right: 0px;
   top: -10px;
}

h2.page-header {
   color: rgb(229, 114, 0);
   margin: 0;
}

p.page-header {
   margin: 0 0 0 30px;
   color: #666;
}

a.back {
   color: #2c3e50;
   ;
   text-decoration: none;
   font-size: 0.9em;
   position: relative;
   top: -10px;
}

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

a.back:hover {
   text-decoration: underline;
}</style>
