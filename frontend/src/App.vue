<template>
  <div id="app">
    <apollo-header/>
    <loading-spinner v-if="authorizing" message="Authorizing"/>
    <router-view v-else></router-view>
    <apollo-footer/>
  </div>
</template>

<script>
import ApolloHeader from './components/ApolloHeader'
import ApolloFooter from './components/ApolloFooter'
import axios from 'axios'

export default {
  components: {
    'apollo-header': ApolloHeader,
    'apollo-footer': ApolloFooter
  },
  data: function () {
    return {
      authorizing: true,
    }
  },
  created: function () {
    if (this.$route.meta.requiresAuth ) {
      axios.get("/api/authenticate").then((response)  =>  {
        localStorage.setItem("user", response.data.firstName+" "+response.data.lastName)
        this.authorizing = false
      }).catch(function (error) {
        self.authorizing = false
        localStorage.removeItem("user")
        router.push({ path: 'unauthorized' })
      })
    } else {
      this.authorizing = false
    }
  }
}
</script>

<style>
body {
  background-color: #002F6C;
  padding: 0;
  margin: 0;
}
</style>
