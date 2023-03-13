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
import { onMounted, onUnmounted, ref, nextTick } from 'vue'
import { useRoute } from 'vue-router'

const collectionsStore = useCollectionsStore()
const route = useRoute()

const toolbar = ref()
const toolbarHeight = ref(0)
const toolbarTop = ref(0)

onMounted( async () => {
   window.addEventListener("scroll", scrollHandler)

   await collectionsStore.getCollectionDetails(route.params.id)
   let targetPID = ""
   if (route.query.item) {
      targetPID = route.query.item
      console.log("TARGET PID "+targetPID)
      collectionsStore.toggleOpen( targetPID )
      nextTick( () => {
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


         let tgtNode = document.getElementById( targetPID )
         if (tgtNode) {
            tgtNode.classList.add("selected")
            scrollToPID(targetPID)
            let extPID = collectionsStore.externalPID(targetPID)
            if ( extPID != "" ) {
               let viewewrDiv = document.getElementById("object-viewer")
               collectionsStore.loadViewer(viewewrDiv, targetPID, extPID)
            }
         }
      })
   }
})

onUnmounted(() => {
   window.removeEventListener("scroll", scrollHandler)
})

function syncView() {
   scrollToPID(collectionsStore.viewerPID)
}

function scrollToPID( pid ) {
   let tgtEle = document.getElementById(pid)
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
