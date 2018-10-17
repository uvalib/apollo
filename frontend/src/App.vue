<template>
  <div id="app">
    <ApolloHeader/>
    <LoadingSpinner v-if="authorizing" message="Authorizing"/>
    <router-view v-else></router-view>
    <ApolloFooter/>
  </div>
</template>

<script>
import ApolloHeader from './components/ApolloHeader'
import ApolloFooter from './components/ApolloFooter'
import axios from 'axios'

export default {
  components: {
    ApolloHeader,
    ApolloFooter
  },
  data: function () {
    return {
      authorizing: true,
    }
  },
  created: function () {
    if (this.$route.meta.requiresAuth ) {
      axios.get("/authenticate").then((response)  =>  {
        localStorage.setItem("user", response.data.firstName+" "+response.data.lastName)
        this.authorizing = false
      }).catch(() => {
        this.authorizing = false
        localStorage.removeItem("user")
        this.$router.push({ path: 'unauthorized' })
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
