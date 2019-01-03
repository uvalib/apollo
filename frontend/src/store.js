import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

// root state object. Holds all of the state for the system
const state = {
  collections: [],
  error: null,
  loading: false
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
  
  collectionsCount: state => {
    return state.collections.length
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
  }
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
}

// // Plugin to listen for error messages being set. After a delay, clear them
// const errorPlugin = store => {
//   store.subscribe((mutation) => {
//     if (mutation.type === "setError") {
//       if ( mutation.payload != null ) {
//         setTimeout( ()=>{ store.commit('setError', null)}, 6000)
//       }
//     }
//   })
// }

// A Vuex instance is created by combining state, getters, actions and mutations
export default new Vuex.Store({
  state,
  getters,
  actions,
  mutations,
  plugins: [] 
})