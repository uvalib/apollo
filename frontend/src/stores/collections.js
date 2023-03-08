import { defineStore } from 'pinia'
import axios from 'axios'

export const useCollectionsStore = defineStore('collections', {
   state: () => ({
      collections: [],
      collectionDetails: {},
      error: null,
      loading: false,
      viewerLoading: false,
      viewerPID: null,
      viewerError: null,
      searchQuery: null,
      selectedPID: ""
   }),
   getters: {
      hasError: state => {
         return state.error != null && state.error != ""
      },
      collectionFound: state => {
         return Object.keys(state.collectionDetails).length > 0
      },
      jsonLink: state => {
         return "/api/collections/" + state.collectionDetails.pid + "?format=json"
      },
      xmlLink: state => {
         return "/api/collections/" + state.collectionDetails.pid + "?format=xml"
      },
      selectedTitle: state => {
         let coll = state.collections.find( c => c.pid == state.selectedPID)
         if ( coll ) {
            return coll.title
         }
         return ""
      }
   },
   actions: {
      collectionSelected(pid) {
         this.selectedPID = pid
      },
      getCollections() {
         this.loading = true
         axios.get("/api/collections").then((response) => {
            if (response.status === 200) {
               this.collections = response.data
            } else {
               this.collections = []
               this.errow = "Internal Error: " + response.data
            }
         }).catch((error) => {
            this.collections = []
            this.error = "Internal Error: " + error
         }).finally(() => {
            this.loading = false
         })
      },

      async getCollectionDetails(pid) {
         this.loading = true
         this.selectedPID = pid
         return axios.get("/api/collections/" + this.selectedPID).then((response) => {
            let model = traverseCollectionDetail(response.data, {})
            this.collectionDetails = model
         }).catch((error) => {
            this.collectionDetails = {}
            if (error.response) {
               this.error = error.response.data
            } else {
               this.error = error
            }
         }).finally(() => {
            this.loading = false
         })
      }
   }
})

// recursive walk of json collection heirarchy into a format more suited
// for presntation. Instead of a list of heirarchical nodes, convert to
// items and attributes. Items contain other items and attributes. Attributes
// are just name/value pairs. Both items and Attributes have some common metadata:
// PID, Type and Sequence.
function traverseCollectionDetail(json, currNode) {
   // init data that is common to all node types (and is single instance):
   // pid, type, sequence and (if present) published date
   commonInit(json, currNode)

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

            if (hasAttribute(currNode.attributes, child.type) === false) {
               // This is the first instance of this node type. Init a blank
               // attribute with no values and add it to the list of attributes for this node
               var attrNode = {}
               commonInit(child, attrNode)
               attrNode.values = []
               currNode.attributes.push(attrNode)
            }

            // Now grab the value and add it to the array of values for the existing attrinute
            var val = { value: child.value }
            if (child.valueURI) {
               val.valueURI = child.valueURI
            }
            attrNode.values.push(val)
         } else {
            // This is another container. Traverse it and append results to children list
            // If this is the first child encountered, create the blank array to hold the children.
            if (!currNode.children) currNode.children = []
            var sub = traverseCollectionDetail(child, {})
            currNode.children.push(sub)
         }
      }
   }
   return currNode
}

// initialize data elements common to both attribue and item nodes
function commonInit(json, currNode) {
   currNode.pid = json.pid
   currNode.type = json.type
   currNode.sequence = json.sequence
}

// See of the list of attributes already includes the attribute spacified
function hasAttribute(attributes, attrType) {
   for (var idx in attributes) {
      let attr = attributes[idx]
      if (attr.type.name == attrType.name) {
         return true
      }
   }
   return false
}