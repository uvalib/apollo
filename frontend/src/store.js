import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

// root state object. Holds all of the state for the system
const state = {
  collections: [],
  collectionDetails: {},
  error: null,
  loading: false,
  viewerLoading: false,
  viewerPID: null, 
  viewerError: null
}

// state getter functions. All are functions that take state as the first param 
// and the getters themselves as the second param. Getter params are passed 
// as a function. Access as a property like: this.$store.getters.NAME
const getters = {
  collections: state => {
    return state.collections
  },
  error: state => {
    return state.error
  },
  hasError: state => {
    return state.error != null && state.error != ""
  },
  isLoading: state => {
    return state.loading
  },
  collectionDetails: state => {
    return state.collectionDetails
  },
  collectionFound: state => {
    return Object.keys(state.collectionDetails).length > 0
  },
  isPublished: state => {
    return state.collectionDetails.publishedAt
  },
  publishedAt: state => {
    return state.collectionDetails.publishedAt
  },
  virgoLink: state => {
    let extPid = ""
    for (var idx in state.collectionDetails.attributes) {
      let attr = state.collectionDetails.attributes[idx]
      if (attr.type.name === "externalPID"){
        extPid = attr.values[0].value
        break
      }
    }
    return process.env.VUE_APP_VIRGO_URL+"/catalog/"+extPid
  },
  jsonLink: state => {
    return "/api/collections/"+state.collectionDetails.pid
  },
  fromSirsi: state=> {
    // A collection is from Sirsi if it has a barcode
    for (var idx in state.collectionDetails.attributes) {
      let attr = state.collectionDetails.attributes[idx]
      if (attr.type.name === "barcode") return true
    }
    return false
  },
  sirsiLink: state => {
    // This should only return a URL for nodes that
    // are top level. A top level node will have a barcode and/or key
    let barcode=""
    let catalogKey = ""
    for (var idx in state.collectionDetails.attributes) {
      let attr = state.collectionDetails.attributes[idx]
      if (attr.type.name === "barcode"){
        barcode = attr.values[0].value
      }
      if (attr.type.name === "catalogKey") {
        catalogKey = attr.values[0].value
      }
    }

    if (barcode.length > 0) {
      return process.env.VUE_APP_SOLR_URL+"/core/select/?q=barcode_facet:"+barcode
    }
    if (catalogKey.length > 0) {
      return process.env.VUE_APP_SOLR_URL+"/core/select/?q=id:"+catalogKey
    }
    return ""
  },
  viewerError: state => {
    return state.viewerError
  },
  viewerPID: state => {
    return state.viewerPID
  },
  isViewerLoading: state => {
    return state.viewerLoading
  }
}

// Synchronous updates to the state. Can be called directly in components like this:
// this.$store.commit('mutation_name') or called from asynchronous actions
const mutations = {
  setError (state, error) {
    state.error = error
  },
  setLoading (state, loading) {
    state.loading = loading
  },
  setCollections(state, colls) {
    if (colls) {
      state.collections = colls
    }
  },
  setCollectionDetails(state, detail) {
    if (detail) {
      state.collectionDetails = detail
    }
  },
  setViewerLoading (state, loading) {
    state.viewerLoading = loading
  },
  setViewerError (state, error) {
    state.viewerError = error
    if (error != "") {
      state.viewerPID = nil
    }
  },
  setViewerPID (state, pid) {
    state.viewerPID = pid
  },
}

// Actions are asynchronous calls that commit mutatations to the state.
// All actions get context as a param which is essentially the entirety of the 
// Vuex instance. It has access to all getters, setters and commit. They are 
// called from components like: this.$store.dispatch('action_name', data_object)
const actions = {
  getCollections( ctx ) {
    ctx.commit('setLoading', true) 
    axios.get("/api/collections").then((response)  =>  {
      if ( response.status === 200) {
        ctx.commit('setCollections', response.data )
      } else {
        ctx.commit('setCollections', []) 
        ctx.commit('setError', "Internal Error: "+response.data) 
      }
    }).catch((error) => {
      ctx.commit('setCollections', []) 
      ctx.commit('setError', "Internal Error: "+error) 
    }).finally(() => {
      ctx.commit('setLoading', false) 
    })
  },

  getCollectionDetails(ctx, collectionID) {
    ctx.commit('setLoading', true) 
    axios.get("/api/collections/"+collectionID).then((response)  =>  {
      let model = traverseCollectionDetail(response.data, {})
      ctx.commit('setCollectionDetails', model)
    }).catch((error) => {
      ctx.commit('setCollectionDetails', {}) 
      if (error.response ) {
        ctx.commit('setError', error.response.data)
      } else {
        ctx.commit('setError', error)
      }
    }).finally(() => {
      ctx.commit('setLoading', false) 
    })  
  }
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
    // Walk children and build attributes and children arrays
    for (var idx in json.children) {
      var child = json.children[idx]
      if (child.type.container === false) {
        // This is an attribute; just grab its value (and valueURI)
        // Important: attributes can be multi-valued. Stuff all values
        // in an array. Init attribues as a blank array of it doesn't exist
        if (!currNode.attributes) currNode.attributes = []

        if  (hasAttribute(currNode.attributes, child.type) === false ) {
          // This is the first instance of this node type. Init a blank
          // attribute with no values and add it to the list of attributes for this node
          var attrNode = {}
          commonInit(child, attrNode)
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
        var sub = traverseCollectionDetail(child, {})
        currNode.children.push( sub )
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
  if (json.publishedAt) {
    currNode.publishedAt = json.publishedAt
  }
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

// A Vuex instance is created by combining state, getters, actions and mutations
export default new Vuex.Store({
  state,
  getters,
  actions,
  mutations,
  plugins: [] 
})