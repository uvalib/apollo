<template>
  <div class="detail-wrapper">
    <apollo-error v-if="error" :message="errorMsg"></apollo-error>
    <template v-else>
      <page-header
        main="Collection Details"
        :sub="title"
        :back="true"></page-header>
      <loading-spinner v-if="loading" message="Loading collection details"/>
      <div v-else class="content">
        <ul id="collection">
          <details-node :model="collection" :depth="0"></details-node>
        </ul>
      </div>
    </template>
  </div>
</template>

<script>
  import axios from 'axios'
  import ApolloError from './ApolloError'
  import CollectionDetailsNode from './CollectionDetailsNode'
  import PageHeader from './PageHeader'

  export default {
    name: 'collection-details',
    components: {
      'apollo-error': ApolloError,
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
        loading: false,
        error: null,
      }
    },
    created: function () {
      this.loading = true;
      var self = this;
      axios.get("/api/collections/"+this.id).then((response)  =>  {
        this.loading = false;
        this.traverseDetails(response.data, this.collection)
        self.error = null
      }).catch(function (error) {
        self.loading = false;
        self.error = true;
        if (error.response) {
          self.errorMsg = error.response.data;
        } else {
          self.errorMsg = "An internal error has occurred"
        }
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
  div.detail-wrapper {
    background-color: white;
    padding: 20px;
  }
</style>
