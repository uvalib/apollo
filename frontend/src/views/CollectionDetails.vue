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
            <CollectionDetailsHeader id="fixed-header" @sync="syncView"/>
            <div class="collection-detail">
               <div class="tree">
                  <div class="toolbar">
                     <span class="toolbar-buttons">
                        <a class="raw" :href="collectionsStore.xmlLink">XML</a>
                        <a class="raw" :href="collectionsStore.jsonLink">JSON</a>
                     </span>
                  </div>
                  <ul class="collection">
                     <CollectionDetailsItem :model="collectionsStore.collectionDetails" :depth="0" :open="true"/>
                  </ul>
               </div>
               <div class="viewer" id="viewer-wrapper">
                  <ApolloError v-if="collectionsStore.viewerError" :message="viewerError"/>
                  <LoadingSpinner v-if="collectionsStore.viewerLoading" message="Loading digital object view"/>
                  <div id="object-viewer"></div>
               </div>
            </div>
         </template>
      </template>
   </div>
</template>

<script setup>
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
const toolbar = ref()
const toolbarHeight = ref(0)
const toolbarTop = ref(0)

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
   setTimeout( () => {
      let tb = document.getElementById("fixed-header")
      if ( tb) {
         toolbar.value = tb
         toolbarHeight.value = tb.offsetHeight
         toolbarTop.value = 0

         // walk the parents of the toolbar and add each top value
         // to find the top of the toolbar relative to document top
         let ele = tb
         if (ele.offsetParent) {
            do {
               toolbarTop.value += ele.offsetTop
               ele = ele.offsetParent
            } while (ele)
         }
         toolbarTop.value -= 5
      }
   }, 1000)
   window.addEventListener("scroll", scrollHandler)
})

onUnmounted(() => {
   window.removeEventListener("scroll", scrollHandler)
})

function syncView() {
   let tgtEle = document.getElementById(collectionsStore.viewerPID)
   if (tgtEle) {
      let nav = document.getElementById("fixed-header")
      var headerOffset = nav.offsetHeight
      var elementPosition = tgtEle.getBoundingClientRect().top
      var offsetPosition = elementPosition - headerOffset
      window.scrollBy({
         top: offsetPosition,
         behavior: "smooth"
      })
   }
}

function scrollHandler( ) {
   if ( toolbar.value) {
      if ( window.scrollY <= toolbarTop.value ) {
         if ( toolbar.value.classList.contains("sticky") ) {
            toolbar.value.classList.remove("sticky")
            let details = document.getElementsByClassName("collection-detail")
            if ( details ) {
               details[0].style.top = `0px`
            }
            let vw = document.getElementById("viewer-wrapper")
            vw.classList.remove("sticky")
            vw.style.top = `0`
            vw.style.left = `0`
         }
      } else {
         if ( toolbar.value.classList.contains("sticky") == false ) {
            let details = document.getElementsByClassName("collection-detail")
            if ( details ) {
               details[0].style.top = `${toolbarHeight.value}px`
            }
            let vw = document.getElementById("viewer-wrapper")
            let currLeft = vw.offsetLeft
            vw.classList.add("sticky")
            vw.style.top = `40px`
            vw.style.left = `${currLeft}px`
            toolbar.value.classList.add("sticky")
         }
      }
   }
}

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

<style scoped lang="scss">
#fixed-header {
   background: white;
   z-index: 1000;
   padding-top:5px;
}
#fixed-header.sticky {
   position: fixed;
   z-index: 1000;
   top: 0;
   left: 20px;
   right: 20px;
}
.collection-detail {
   display: flex;
   flex-flow: row nowrap;
   position: relative;

   div.tree {
      width: 38%;
      min-width: 600px;
      .toolbar {
         font-size: 0.8em;
         position: relative;
         margin: 0 0 0 20px;
         border-bottom: 1px solid #ccc;
         padding: 10px 0;
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-end;
         span.toolbar-buttons {
            display: inline-block;
         }
      }
      ul.collection {
         margin-top: 0;
         -webkit-padding-start: 20px;
      }
   }
   .viewer {
      width: 60%;
      min-width: 800px;
      position: relative;
      div#object-viewer {
         padding: 5px 25px;
      }
   }
   #viewer-wrapper {
      margin-left: 20px;
   }
   #viewer-wrapper.sticky {
      position: fixed;
      z-index: 1000;
      right: 40px;
      bottom: 0;

   }
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

div.detail-wrapper {
   background-color: white;
   padding: 20px;
}
</style>
