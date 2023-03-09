import { defineStore } from 'pinia'
import axios from 'axios'
const CURIO_URL = import.meta.env.VITE_CURIO_URL

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
      },
      externalPID: state => {
         return (pid) => {
            let extPID = ""
            let node = findNode(state.collectionDetails, pid)
            if (node) {
               node.attributes.some( na => {
                  if (na.type.name == "externalPID") {
                     extPID = na.values[0].value
                  }
                  return extPID != ""
               })
            }
            return extPID
         }
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
            model.open = true
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
      },

      loadViewer(viewewrDiv, pid, externalID) {
         // generate the oembedURI and request embedding info
         // https://curio.lib.virginia.edu/oembed?url=https%3A%2F%2Fcurio.lib.virginia.edu%2Fview%2Fuva-lib%3A2528443
         this.viewerLoading = true
         let qp = encodeURIComponent(CURIO_URL+"/view/"+externalID)
         let oembedUri = CURIO_URL+"/oembed?url="+qp

         axios.get(oembedUri).then((response)  =>  {
            // set a global flag to make the browser think JS jas not all been
            // loaded. Without this, the JS file included in the response
            // will not load, and the viewer will not render
            window.embedScriptIncluded = false

            viewewrDiv.innerHTML = response.data.html
            this.viewerPID = pid
         }).catch((error) => {
            if ( error.message ) {
               this.viewerError = error.message
            } else {
               this.viewerError = error.response.data
            }
         }).finally(() => {
            this.viewerLoading = false
         })
      },

      toggleOpen( pid ) {
         toggleNodeOpen(this.collectionDetails, pid)
      },
      closeAll() {
         closeAllOpenNodes(this.collectionDetails)
      }
   },
})

function findNode( currNode, pid ) {
   let out = null
   if ( currNode.pid == pid) {
      currNode.open = !currNode.open
      return currNode
   }
   if (currNode.children) {
      currNode.children.some( node => {
         if (node.children ) {
            out = findNode(node, pid)
         } else {
            if ( node.pid == pid) {
               out = node
            }
         }
         return out != null
      })
   }
   return out
}

function toggleNodeOpen( currNode, pid ) {
   if ( currNode.pid == pid) {
      currNode.open = !currNode.open
      return true
   }

   if (currNode.children) {
      let done = false
      currNode.children.some( node => {
         if (node.children ) {
            if (toggleNodeOpen(node, pid)) {
               done = true
            }
         } else {
            if ( node.pid == pid) {
               currNode.open = !currNode.open
            }
         }
         return done == true
      })
   }
   return false
}

function closeAllOpenNodes( currNode ) {
   if (currNode.children) {
      currNode.open = false
      currNode.children.forEach( node => {
         if (node.children ) {
            closeAllOpenNodes(node)
         }
      })
   }
   return false
}

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
      currNode.open = false

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