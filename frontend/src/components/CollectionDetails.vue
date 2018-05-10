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
          <div id="object-viewer">
            <p id="view-placeholder" class="hint">Click 'View Digital Object' from the tree on the left to view it here.</p>
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
        errorMsg: null
      }
    },
    created: function () {
      axios.get("/api/collections/"+this.id).then((response)  =>  {
        this.loading = false;
        this.traverseDetails(response.data, this.collection)
      }).catch((error) => {
        this.loading = false
        this.errorMsg =  error.response.data
      });
    },
    methods: {
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
  h4.do-header {
    margin: 0;
    border-bottom: 1px solid #ccc;
    padding-bottom: 10px;
    margin-left: 20px;
    margin-bottom: 15px;
  }
  p.hint {
    color: #999;
    font-style: italic;
  }
  div.detail-wrapper {
    background-color: white;
    padding: 20px;
  }
  ul.collection {
    margin-top:0;
  }
</style>
