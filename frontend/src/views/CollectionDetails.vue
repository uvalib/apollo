<template>
   <div class="detail-wrapper">
      <ApolloError v-if="collectionsStore.hasError" :message="collectionsStore.error" />
      <template v-else>
         <PageHeader main="Collection Details" :sub="collectionsStore.selectedTitle" />
         <LoadingSpinner v-if="collectionsStore.loading" message="Loading collection details" />
         <div v-else-if="collectionsStore.collectionFound === false" class="content">
            <h4>No data found!</h4>
         </div>
         <template v-else>
            <CollectionDetailsHeader id="fixed-header"/>
            <div class="content collection-detail">
               <div class="pure-u-9-24">
                  <div class="toolbar">
                  <span class="toolbar-buttons">
                     <a class="raw" :href="collectionsStore.xmlLink">XML</a>
                     <a class="raw" :href="collectionsStore.jsonLink">JSON</a>
                  </span>
                  </div>
                  <ul class="collection">
                     <CollectionDetailsItem :model="collectionsStore.collectionDetails" :depth="0"/>
                  </ul>
               </div>
               <div class="pure-u-15-24">
                  <ApolloError v-if="collectionsStore.viewerError" :message="viewerError"/>
                  <div v-else id="viewer-wrapper">
                  <LoadingSpinner v-if="collectionsStore.viewerLoading" message="Loading digital object view"/>
                  <div id="object-viewer"></div>
                  </div>
               </div>
            </div>
         </template>
      </template>
   </div>
</template>

<script setup>
import moment from 'moment'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import CollectionDetailsHeader from '@/components/CollectionDetailsHeader.vue'
import CollectionDetailsItem from '@/components/CollectionDetailsItem.vue'
import PageHeader from '@/components/PageHeader.vue'
import ApolloError from '@/views/ApolloError.vue'
import { useCollectionsStore } from '@/stores/collections'
import { onBeforeMount, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'

const collectionsStore = useCollectionsStore()
const route = useRoute()

const targetAncestry = ref([])
const targetPID = ref("")

onBeforeMount( async () => {
   await collectionsStore.getCollectionDetails(route.params.id)
   if (route.query.item) {
      targetPID.value = route.query.item

      // A target was specified; get a list of ancestor nodes
      // that can be used to expand the display tree to that item
      getAncestry(collectionStore.collectionDetails)
      targetAncestry.value.reverse()

      // NOTE: for the first case, need to wait for one more node; the root itself
      // mount events come in for each child followed by one for root node
      targetAncestry.value[0].childNodeCount += 1
   }
})
onMounted(() => {
   // setup sticky toolbar
})

onUnmounted(() => {
   // cleanup sticky
})

      // formatDateTime: function( ts ) {
      //   let m = moment(ts, "YYYY-MM-DDTHH:mm:ssZ")
      //   return m.utcOffset("+0000").format("YYYY-MM-DD hh:mma")
      // },

// Walk the collection data and find the targetPID specified by the query params
// Populate an array of targetAncestry data including node counts for the relevant tree branches
const getAncestry = ((currNode) => {
   if ( currNode.pid === targetPID.value ) {
      // its a match. Return true to start unwinding the recursion
      return true
   } else {
      // Nope; walk the child nodes...
      for (var idx in currNode.children) {
         if ( getAncestry(currNode.children[idx]) ) {
            // the target was hit in this branch. Add to ancestors and exit
            targetAncestry.value.push( {pid: currNode.pid, childNodeCount: currNode.children.length} )
            return true
         }
      }
   }
   // Child branch traversed with no hits; return false
   return false
})

      // handleNodeMounted: function() {
      //   // only care about this event if a target was specified in the query params
      //   if (props.targetPID && this.targetAncestry.length > 0) {
      //     // Wait for each child of the targetAncestry node to be mounted
      //     this.targetAncestry[0].childNodeCount -= 1
      //     if ( this.targetAncestry[0].childNodeCount <= 0) {
      //       // Once all are mounted, toss the head of the list and
      //       // open the next ancestor - or scroll to target if all are open
      //       this.targetAncestry.shift()
      //       if ( this.targetAncestry.length == 0) {
      //         this.scrollToTarget()
      //       } else {
      //         EventBus.$emit("expand-node", this.targetAncestry[0].pid)
      //       }
      //     }
      //   }
      // },

// const scrollToTarget = (() => {
//    let ele = $("li#"+props.targetPID)

//    /// if this item has a digital object, click the ciew button to show it
//    let doViewerBtn = ele.find("span.do-button")
//    if ( doViewerBtn.length > 0) {
//       doViewerBtn.trigger("click")
//    } else {
//       ele.addClass("target")
//       setTimeout(function() {
//       ele.removeClass("target")
//       },1000)
//    }

//    $([document.documentElement, document.body]).animate({
//       scrollTop: ele.offset().top-$(".fixed-header").outerHeight(true)
//    }, 100);
// })
</script>

<style scoped>
.toolbar {
   font-size: 0.8em;
   position: relative;
   margin: 0 0 0 20px;
   border-bottom: 1px solid #ccc;
   padding: 10px 0;
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-end;
}

.publication {
   margin-top: 2px;
   font-size: 0.9em;
}

.publication .label {
   font-weight: bold;
   margin-right: 10px;
}

.raw {
   border-radius: 10px;
   background: rgb(240, 130, 40);
   opacity: 0.75;
   cursor: pointer;
   text-decoration: none;
   font-weight: 500;
   color: white;
   padding: 3px 15px;
   border-radius: 10px;
   cursor: pointer;
   margin-left: 5px;
}

span.toolbar-buttons {
   display: inline-block;
}

span.publish {
   background: #3a3;
}

div#object-viewer {
   padding: 5px 20px;
}

#viewer-tools {
   text-align: right;
   margin: 15px 0;
}

div.detail-wrapper {
   background-color: white;
   padding: 20px;
}

ul.collection {
   margin-top: 0;
   -webkit-padding-start: 20px;
}</style>
