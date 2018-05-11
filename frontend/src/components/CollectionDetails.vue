<template>
  <div class="detail-wrapper">
    <apollo-error v-if="errorMsg" :message="errorMsg"></apollo-error>
    <template v-else>
      <page-header
        main="Collection Details"
        :sub="title"
        :back="true"></page-header>
      <loading-spinner v-if="loading" message="Loading collection details"/>
      <div v-else-if="Object.keys(collection).length === 0" class="content">
        <h4>No data found!</h4>
      </div>
      <div v-else class="content pure-g">
        <div class="pure-u-9-24">
          <h4 class="do-header">Collection Structure</h4>
          <ul class="collection">
            <details-node :model="collection" :depth="0"></details-node>
          </ul>
        </div>
        <div class="pure-u-15-24">
          <h4 class="do-header">Digitial Object Viewer</h4>
          <apollo-error v-if="viewerError" :message="viewerError"></apollo-error>
          <div v-else id="viewer-wrapper">
            <div id="object-viewer">
              <p id="view-placeholder" class="hint">Click 'View Digital Object' from the tree on the left to view it here.</p>
            </div>
            <div v-if="viewerVisible" id="viewer-tools">
              <!-- <a class="do-button" href="#" target="_blank">PDF</a> -->
              <a class="do-button" :href="iiifManufestURL()" target="_blank">IIIF Manifest</a>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
  import axios from 'axios'
  import CollectionDetailsNode from './CollectionDetailsNode'
  import PageHeader from './PageHeader'
  import EventBus from './EventBus'

  export default {
    name: 'collection-details',
    components: {
      'details-node': CollectionDetailsNode,
      'page-header': PageHeader
    },
    props: {
      id: String,
      title: String
    },
    data: function () {
      return {
        collection: {},
        loading: true,
        viewerVisible: false,
        errorMsg: null,
        viewerError: null,
        activePID: ""
      }
    },
    created: function () {
      axios.get("/api/collections/"+this.id).then((response)  =>  {
        this.loading = false
        this.traverseDetails(response.data, this.collection)
      }).catch((error) => {
        this.loading = false
        this.errorMsg =  error.response.data
      }).finally(() => {
        this.loading = false
      })
    },
    mounted: function (){
      EventBus.$on("viewer-clicked", this.handleViewerClicked)
      EventBus.$on("viewer-opened", this.handleViewerOpened)
      EventBus.$on("viewer-error", this.handleViewerError)
    },
    methods: {
      handleViewerClicked: function() {
        this.viewerError = null
        this.activePID = ""
      },
      handleViewerOpened: function(pid) {
        this.viewerVisible = true
        this.activePID = pid
      },
      handleViewerError: function(msg) {
        this.viewerError = msg
        this.activePID = ""
      },
      iiifManufestURL: function() {
        return "https://tracksys.lib.virginia.edu:8080/"+this.activePID
      },
      traverseDetails: function(json, currNode) {
        // every node has at least a PID and name obj (pid, name)
        currNode.pid = json.pid
        currNode.name = json.name

        // If it does not have a corresponding VALUE attribute, it is a container node
        // container nodes contain attributes (simple name/value pairs) and children.
        // NOTE: All container nodes should have a title attribute
        // non-container nodes just contain attributes.
        //     Ex: in our mountain work, issue is not a container. It has 2 attributes;
        //         one for issue title and another a digitalObject with a link to the oEnbed viewer
        if (json.value) {
          // This node has a value; it is an attribute. just poulate value
          currNode.value = json.value
          if (json.valueURI) {
            currNode.valueURI = json.valueURI
          }
        } else {
            // This node has no value so it is a container.
            // Walk children and build attributes and children arrays
            for (var idx in json.children) {
              var child = json.children[idx]
              if (child.value) {
                // This is an attribute traverse its detail and add it to the attributes list
                if (!currNode.attributes) currNode.attributes = []
                var sub = this.traverseDetails(child, {})
                currNode.attributes.push( sub )
              } else {
                // This is another container. Traverse it and append results to children list
                // If this is the first child encountered, create the blank array to hold the children.
                if (!currNode.children) currNode.children = []
                var sub = this.traverseDetails(child, {})
                currNode.children.push( sub )
              }
            }
        }
        // The node, children and attributes have been populated; return result
        return currNode
      }
    }
  }
</script>

<style scoped>
  div#object-viewer {
    padding: 0px 20px;
  }
  #viewer-tools {
    text-align: right;
    margin: 15px 0;
  }
  a.do-button {
    padding: 5px 25px 4px 25px;
    border-radius: 15px;
    background: #0078e7;
    color: white;
    opacity: 0.7;
    cursor: pointer;
    text-decoration: none;
    margin-left: 5px;
    font-size: 0.9em;
  }
  a.do-button:hover {
    opacity: 1;
  }
  h4.do-header {
    margin: 0;
    border-bottom: 1px solid #ccc;
    padding-bottom: 10px;
    margin-left: 20px;
    margin-bottom: 15px;
  }
  p.hint {
    color: #999;
    margin: 25px 0;
    text-align: center;;
    /* font-style: italic; */
  }
  div.detail-wrapper {
    background-color: white;
    padding: 20px;
  }
  ul.collection {
    margin-top:0;
  }
</style>
