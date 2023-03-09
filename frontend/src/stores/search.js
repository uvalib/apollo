import { defineStore } from 'pinia'
import axios from 'axios'

export const useSearchStore = defineStore('search', {
   state: () => ({
      searching: true,
      errorMsg: "",
      searchResults: null,
      query: ""
   }),
	getters: {
   },
	actions: {
      search( ) {
         if ( this.query == "") return
         this.searching = true
         axios.get("/api/search?q=" + this.query).then((response) => {
            this.searchResults = response.data
         }).catch((error) => {
            if (error.response) {
               this.errorMsg = error.response.data
            } else {
               this.errorMsg = error
            }
         }).finally(() => {
            this.searching = false
         })
      }
   }
})