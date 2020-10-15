<template>
  <div class="detail-wrapper">
    <ApolloError v-if="hasError" :message="error"/>
    <template v-else>
      <PageHeader
        main="Collection Details"
        :sub="title"
      />
      <LoadingSpinner v-if="isLoading" message="Loading collection details"/>
      <div v-else-if="collectionFound === false" class="content">
        <h4>No data found!</h4>
      </div>
      <template v-else>
        <!-- content header; this portion is fixed and wont scroll offscreen -->
        <CollectionDetailsHeader/>

        <!-- main content; this will scroll -->
        <div class="content pure-g collection-detail">
          <div class="pure-u-9-24">
            <div class="toolbar">
              <span class="toolbar-buttons">
                <a class="raw" :href="jsonLink" target="_blank">JSON</a>
              </span>
            </div>
            <ul class="collection">
              <CollectionDetailsItem :model="collectionDetails" :depth="0"/>
            </ul>
          </div>
          <div class="pure-u-15-24">
            <ApolloError v-if="viewerError" :message="viewerError"/>
            <div v-else id="viewer-wrapper">
              <LoadingSpinner v-if="isViewerLoading" message="Loading digital object view"/>
              <div id="object-viewer"></div>
            </div>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<script>
  import moment from 'moment'
  import LoadingSpinner from '@/components/LoadingSpinner'
  import CollectionDetailsHeader from '@/components/CollectionDetailsHeader'
  import CollectionDetailsItem from '@/components/CollectionDetailsItem'
  import PageHeader from '@/components/PageHeader'
  import EventBus from '@/components/EventBus'
  import PinnedScroll from '@/components/PinnedScroll'
  import ApolloError from '@/views/ApolloError'
  import { mapGetters } from 'vuex'

  export default {
    name: 'CollectionDetails',
    components: {
      ApolloError,
      LoadingSpinner,
      CollectionDetailsHeader,
      CollectionDetailsItem,
      PageHeader
    },

    computed: {
      ...mapGetters([
        'isLoading',
        'hasError',
        'error',
        'collectionDetails',
        'collectionFound',
        'jsonLink',
        'isViewerLoading',
        'viewerError',
      ])
    },

    props: {
      id: String,
      title: String,
      targetPID: String
    },

    data: function () {
      return {
        targetAncestry: [],
        pinHeader: new PinnedScroll("div.fixed-header", 216),
        pinViewer: new PinnedScroll("#viewer-wrapper", 210, "div.fixed-header")
      }
    },

    created: function () {
      this.$store.dispatch('getCollectionDetails', this.id).then(()  =>  {
        if (this.targetPID) {
          // A target was specified; get a list of ancestor nodes
          // that can be used to expand the display tree to that item
          this.getAncestry(this.collectionDetails)
          this.targetAncestry.reverse()

          // NOTE: for the first case, need to wait for one more node; the root itself
          // mount events come in for each child followed by one for root node
          this.targetAncestry[0].childNodeCount+=1
        }
      })
    },

    mounted: function (){
      EventBus.$on('node-mounted', this.handleNodeMounted)
      this.pinHeader.register()
      this.pinViewer.register()
    },

    destroyed() {
      this.pinHeader.unregister()
      this.pinViewer.unregister()
    },

    methods: {
      formatDateTime: function( ts ) {
        let m = moment(ts, "YYYY-MM-DDTHH:mm:ssZ")
        return m.utcOffset("+0000").format("YYYY-MM-DD hh:mma")
      },

      // Walk the collection data and find the targetPID specified by the query params
      // Populate an array of targetAncestry data including node counts for the relevant tree branches
      getAncestry: function(currNode) {
        if ( currNode.pid === this.targetPID ) {
          // its a match. Return true to start unwinding the recursion
          return true
        } else {
          // Nope; walk the child nodes...
          for (var idx in currNode.children) {
            if ( this.getAncestry(currNode.children[idx]) ) {
              // the target was hit in this branch. Add to ancestors and exit
              this.targetAncestry.push( {pid: currNode.pid, childNodeCount: currNode.children.length} )
              return true
            }
          }
        }
        // Child branch traversed with no hits; return false
        return false
      },

      handleNodeMounted: function() {
        // only care about this event if a target was specified in the query params
        if (this.targetPID && this.targetAncestry.length > 0) {
          // Wait for each child of the targetAncestry node to be mounted
          this.targetAncestry[0].childNodeCount -= 1
          if ( this.targetAncestry[0].childNodeCount <= 0) {
            // Once all are mounted, toss the head of the list and
            // open the next ancestor - or scroll to target if all are open
            this.targetAncestry.shift()
            if ( this.targetAncestry.length == 0) {
              this.scrollToTarget()
            } else {
              EventBus.$emit("expand-node", this.targetAncestry[0].pid)
            }
          }
        }
      },

      scrollToTarget: function() {
        let ele = $("li#"+this.targetPID)

        /// if this item has a digital object, click the ciew button to show it
        let doViewerBtn = ele.find("span.do-button")
        if ( doViewerBtn.length > 0) {
          doViewerBtn.trigger("click")
        } else {
          ele.addClass("target")
          setTimeout(function() {
            ele.removeClass("target")
          },1000)
        }

        $([document.documentElement, document.body]).animate({
          scrollTop: ele.offset().top-$(".fixed-header").outerHeight(true)
        }, 100);
      }
    }
  }
</script>

<style scoped>
  .toolbar  {
    font-size: 0.8em;
    position: relative;
    text-align: right;
    margin: 0 0 0 20px;
    border-bottom: 1px solid #ccc;
    padding: 10px 0;
  }
  .publication {
    margin-top: 2px;
    font-size: 0.9em;
  }
  .publication .label {
    font-weight: bold;
    margin-right: 10px;
  }
  .sirsi, .raw, .publish, .virgo  {
    border-radius: 10px;
    background: #0078e7;
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
  .sirsi:hover, .raw:hover, .publish:hover {
    opacity: 1;
  }
  .virgo {
    padding: 2px 10px;
  }
  .raw {
    background: rgb(240, 130, 40);
  }
  span.toolbar-buttons {
    margin-bottom: 8px;
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
    margin-top:0;
    -webkit-padding-start: 20px;
  }
</style>
