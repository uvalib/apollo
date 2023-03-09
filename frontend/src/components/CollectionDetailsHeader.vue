<template>
   <div class="content">
      <h4 class="do-header">
         <span>Collection Structure</span>
         <span class="helper-buttons">
            <span class="helper-icon top" @click="scrollTopClick" title="Scroll to top"></span>
            <span class="helper-icon collapse" @click="collapseClick" title="Collapse all"></span>
         </span>
      </h4>

      <h4 class="do-header pad-left">
         <span>Digitial Object Viewer</span>
         <span v-if='!collectionsStore.viewerPID' class="hint">
            Click 'View Digital Object' from the tree on the left to view it below
         </span>
         <span v-else class="helper-buttons">
            <span class="helper-icon sync" @click="syncClick" title="Sync Tree"></span>
         </span>
      </h4>
   </div>
</template>

<script setup>
import { useCollectionsStore } from '@/stores/collections'

const collectionsStore = useCollectionsStore()

const emit = defineEmits(['sync'])

const scrollTopClick = (() => {
   var scrollStep = -window.scrollY / (500 / 10),
   scrollInterval = setInterval(()=> {
      if ( window.scrollY != 0 ) {
         window.scrollBy( 0, scrollStep )
      } else {
         clearInterval(scrollInterval)
      }
   },10)
})
const collapseClick = (() => {
   collectionsStore.closeAll()
})
const syncClick = (() => {
   emit("sync")
})
</script>

<style scoped lang="scss">
.content {
   display: flex;
   flex-flow: row nowrap;
   justify-content: space-between;
   padding-left: 20px;

   h4.do-header {
      margin: 0;
      border-bottom: 1px solid #ccc;
      padding-bottom: 10px;
      margin-bottom: 0px;
      flex-grow: 1;
   }

   h4.do-header.pad-left {
      margin-left: 20px;
   }

   .hint {
      color: #999;
      margin: 0;
      text-align: right;
      font-size: 0.85em;
      float: right;
      font-weight: 500;
   }

   .helper-icon {
      display: inline-block;
      width: 20px;
      height: 20px;
      opacity: 0.3;
      cursor: pointer;
      margin-left: 5px;
   }

   .helper-icon.collapse {
      background-image: url(../assets/collapse.png);
   }

   .helper-icon.top {
      background-image: url(../assets/top.png);
   }

   .helper-icon.sync {
      background-image: url(../assets/sync.png);
   }

   .helper-buttons {
      float: right;
   }

   .helper-icon:hover {
      opacity: 0.8;
   }
}</style>
