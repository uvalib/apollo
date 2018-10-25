<template>
  <div class="detail-wrapper">
    <ApolloError v-if="errorMsg" :message="errorMsg"/>
    <template v-else>
      <PageHeader
        main="Collection Details"
        :sub="title"
        :back="true"/>
      <LoadingSpinner v-if="loading" message="Loading collection details"/>
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
            <a class="raw" :href="jsonLink" target="_blank">JSON</a>
            <a v-if="hasBarcode" class="sirsi" :href="sirsiLink" target="_blank">Sirsi</a>
            <div v-if="published" class="publication">
              <span class="label">Last Published:</span>
              <span class="date">{{ formattedPublishedDate }}</span>
              <a class="virgo" :href="virgoLink" target="_blank">Virgo</a>
            </div>
          </div>

          <ul class="collection">
            <CollectionDetailsNode :model="collection" :depth="0"/>
          </ul>
        </div>

        <div class="pure-u-15-24">
          <h4 class="do-header">Digitial Object Viewer</h4>
          <ApolloError v-if="viewerError" :message="viewerError"/>
          <div v-else id="viewer-wrapper">
            <div id="object-viewer">
              <p id="view-placeholder" class="hint">Click 'View Digital Object' from the tree on the left to view it here.</p>
            </div>
            <div v-if="viewerVisible" id="viewer-tools">
              <a v-if="iiifAvailable" class="do-button" :href="iiifManufestURL" target="_blank">IIIF Manifest</a>
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
  import LoadingSpinner from '@/components/LoadingSpinner'
  import CollectionDetailsNode from '@/components/CollectionDetailsNode'
  import PageHeader from '@/components/PageHeader'
  import EventBus from '@/components/EventBus'

  export default {
    components: {
      LoadingSpinner,
      CollectionDetailsNode,
      PageHeader
    },

    props: {
      id: String,
      title: String,
      targetPID: String
    },

    data: function () {
      return {
        collection: {},
        loading: true,
        viewerVisible: false,
        errorMsg: null,
        viewerError: null,
        activePID: "",
        ancestry: []
      }
    },

    computed: {
      iiifAvailable: function() {
        // as of 7/2018 all items in Apollo that have barcodes are images
        // and have IIIF available. WSLS does not have barcoces, and is not
        // image-based. Easy check for now;
        let iiif = this.hasBarcode
        return iiif
      },
      hasBarcode: function() {
        for (var idx in this.collection.attributes) {
          let attr = this.collection.attributes[idx]
          if (attr.type.name === "barcode") return true
        }
        return false
      },
      published: function() {
        return this.collection.publishedAt
      },
      formattedPublishedDate: function() {
        let m = moment(this.collection.publishedAt, "YYYY-MM-DDTHH:mm:ssZ")
        return m.utcOffset("+0000").format("YYYY-MM-DD hh:mma")
      },
      jsonLink: function() {
        return "/api/collections/"+this.collection.pid
      },
      sirsiLink: function() {
        // This should only return a URL for nodes that
        // are top level. A top level node will have a barcode and/or key
        let barcode=""
        let catalogKey = ""
        for (var idx in this.collection.attributes) {
          let attr = this.collection.attributes[idx]
          if (attr.type.name === "barcode"){
            barcode = attr.values[0].value
          }
          if (attr.type.name === "catalogKey") {
            catalogKey = attr.values[0].value
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
      virgoLink: function() {
        let extPid = ""
        for (var idx in this.collection.attributes) {
          let attr = this.collection.attributes[idx]
          if (attr.type.name === "externalPID"){
            extPid = attr.values[0].value
            break
          }
        }
        return "http://search.lib.virginia.edu/catalog/"+extPid
      },
      iiifManufestURL: function() {
        return "https://tracksys.lib.virginia.edu:8080/"+this.activePID+"/manifest.json"
      }
    },

    created: function () {
      axios.get("/api/collections/"+this.id).then((response)  =>  {
        // parse json tree response into the collection model
        this.traverseDetails(response.data, this.collection)
        if (this.targetPID) {
          // A target was specified; get a list of ancestor nodes
          // that can be used to expand the display tree to that item
          this.getAncestry(this.collection)
          this.ancestry.reverse()

          // NOTE: for the first case, need to wait for one more node; the root itself
          // mount events come in for each child followed by one for root node
          this.ancestry[0].childNodeCount+=1
        }
      }).catch((error) => {
        if (error.response ) {
          this.errorMsg =  error.response.data
        } else {
          this.errorMsg =  error
        }
      }).finally(() => {
        this.loading = false
      })
    },

    mounted: function (){
      EventBus.$on("viewer-clicked", this.handleViewerClicked)
      EventBus.$on("viewer-opened", this.handleViewerOpened)
      EventBus.$on("viewer-error", this.handleViewerError)
      EventBus.$on('node-mounted', this.handleNodeMounted)
    },

    methods: {
      // Walk the collection data and find the targetPID specified by the query params
      // Populate an array of ancestry data including node counts for the relevant tree branches
      getAncestry: function(currNode) {
        if ( this.targetPID.length == 0) return;

        if ( currNode.pid === this.targetPID ) {
          // its a match. Return true to start unwinding the recursion
          return true
        } else {
          // Nope; walk the child nodes...
          for (var idx in currNode.children) {
            if ( this.getAncestry(currNode.children[idx]) ) {
              // the target was hit in this branch. Add to ancestors and exit
              this.ancestry.push( {pid: currNode.pid, childNodeCount: currNode.children.length} )
              return true
            }
          }
        }
        // Child branch traversed with no hits; return false
        return false
      },

      handleNodeMounted: function() {
        // only care about this event if a target was specified in the query params
        if (this.targetPID && this.ancestry.length > 0) {
          // Wait for each child of the ancestry node to be mounted
          this.ancestry[0].childNodeCount -= 1
          if ( this.ancestry[0].childNodeCount <= 0) {
            // Once all are mounted, toss the head of the list and
            // open the next ancestor - or scrolll to target if all are open
            this.ancestry.shift()
            if ( this.ancestry.length == 0) {
              this.scrollToTarget()
            } else {
              EventBus.$emit("expand-node", this.ancestry[0].pid)
            }
          }
        }
      },

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

      publishClicked: function() {
        let resp = confirm("Publish this collection?")
        if (!resp) return

        axios.post("/api/publish/"+this.collection.pid).then(()  =>  {
          alert("The publication process has been started. The collection will appear in Virgo within 24 hours.")
        }).catch((error) => {
          alert("Unable to publish collection: "+error.response)
        })
      },

      // initialize data elements common to both attribue and container nodes
      commonInit: function(json, currNode) {
        currNode.pid = json.pid
        currNode.type = json.type
        currNode.sequence = json.sequence
        if (json.publishedAt) {
          currNode.publishedAt = json.publishedAt
        }
      },

      // See of the list of attributes already includes an attribute of
      // the target type specified in tgtType
      hasAttribute: function(attributes, tgtType) {
        for (var idx in attributes) {
          let attr = attributes[idx]
          if (attr.type.name == tgtType.name) {
            return true
          }
        }
        return false
      },

      traverseDetails: function(json, currNode) {
        // init data that is common to all node types (and is single instance):
        // pid, type, sequence and (if present) published date
        this.commonInit(json,currNode)

        // Detect and handle container nodes differently; recursively walk their children.
        // Container nodes contain attributes (simple name/value pairs) and children.
        // Examples of containers are collection, year and issue.
        if (json.type.container === true) {
          // Walk children and build attributes and children arrays
          for (var idx in json.children) {
            var child = json.children[idx]
            if (child.type.container === false) {
              // This is an attribute; just grab its value (and valueURI)
              // Important: attributes can be multi-valued. Stuff all values
              // in an array. Init attribues as a blank array of it doesn't exist
              if (!currNode.attributes) currNode.attributes = []

              if  (this.hasAttribute(currNode.attributes, child.type) === false ) {
                // This is the first instance of this node type. Init a blank
                // attribute with no values and add it to the list of attributes for this node
                var attrNode = {}
                this.commonInit(child, attrNode)
                attrNode.values = []
                currNode.attributes.push( attrNode )
              }

              // Now grab the value and add it to the array of values for the existing attrinute
              var val = {value: child.value}
              if (child.valueURI) {
                val.valueURI = child.valueURI
              }
              attrNode.values.push(val)
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
      },

      scrollToTarget: function() {
        let ele = $("li#"+this.targetPID)
        let doViewerBtn = ele.find("td.pure-button.dobj")
        doViewerBtn.trigger("click")

        $([document.documentElement, document.body]).animate({
          scrollTop: ele.offset().top-5
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
