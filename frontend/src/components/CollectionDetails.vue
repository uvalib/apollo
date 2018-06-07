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
          <h4 class="do-header">
            <span>Collection Structure</span>
          </h4>

          <div class="toolbar">
            <span @click="publishClicked" class="publish">Publish Collection</span>
            <a class="raw" :href="jsonLink()" target="_blank">JSON</a>
            <a class="sirsi" :href="sirsiLink()" target="_blank">Sirsi</a>
            <div v-if="published()" class="publication">
              <span class="label">Last Published:</span>
              <span class="date">{{ formattedPublishedDate() }}</span>
              <a class="virgo" :href="virgoLink()" target="_blank">Virgo</a>
            </div>
          </div>

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
  import moment from 'moment'
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

      published: function() {
        return this.collection.publishedAt
      },

      formattedPublishedDate: function() {
        let m = moment(this.collection.publishedAt, "YYYY-MM-DDTHH:mm:ssZ")
        return m.utcOffset("+0000").format("YYYY-MM-DD hh:mma")
      },

      virgoLink: function() {
        let extPid = ""
        for (var idx in this.collection.attributes) {
          let attr = this.collection.attributes[idx]
          if (attr.type.name === "externalPID"){
            extPid = attr.value
            break
          }
        }
        return "http://search.lib.virginia.edu/catalog/"+extPid
      },

      jsonLink: function() {
        // This should only return a URL for nodes that
        // are top level. A top level node will have a barcode and/or key
        for (var idx in this.collection.attributes) {
          let attr = this.collection.attributes[idx]
          if (attr.type.name === "barcode" || attr.type.name === "catalogKey") {
            return "/api/collections/"+this.collection.pid
          }
        }
        return ""
      },

      sirsiLink: function(model) {
        // This should only return a URL for nodes that
        // are top level. A top level node will have a barcode and/or key
        let barcode=""
        let catalogKey = ""
        for (var idx in this.collection.attributes) {
          let attr = this.collection.attributes[idx]
          if (attr.type.name === "barcode"){
            barcode = attr.value
          }
          if (attr.type.name === "catalogKey") {
            catalogKey = attr.value
          }
        }

        if (barcode.length > 0) {
          return "http://solr.lib.virginia.edu:8082/solr/core/select/?q=barcode_facet:"+barcode
        }
        if (catalogKey.length > 0) {
          return "http://solr.lib.virginia.edu:8082/solr/core/select/?q=id:"+catalogKey
        }
        return ""
      },

      publishClicked: function() {
        let resp = confirm("Publish this collection?")
        if (!resp) return

        axios.post("/api/publish/"+this.collection.pid).then((response)  =>  {
          alert("The publication process has been started. The collection will appear in Virgo within 24 hours.")
        }).catch((error) => {
          alert("Unable to publish collection: "+error.response)
        })
      },

      iiifManufestURL: function() {
        return "https://tracksys.lib.virginia.edu:8080/"+this.activePID+"/manifest.json"
      },

      traverseDetails: function(json, currNode) {
        // every node has at least a PID and type obj (pid, name)
        currNode.pid = json.pid
        currNode.type = json.type
        currNode.sequence = json.sequence
        if (json.publishedAt) {
          currNode.publishedAt = json.publishedAt
        }

        // If it does not have a corresponding VALUE attribute, it is a container node.
        // Container nodes contain attributes (simple name/value pairs) and children.
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
        return currNode
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
    padding: 0 5px 8px 0;
  }
  .publication {
    margin-top: 12px;
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
  span.publish {
    background: #3a3;
  }
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
    -webkit-padding-start: 20px;
  }
</style>
