<template>
  <div id="app">
    <apollo-header/>
    <router-view></router-view>
    <apollo-footer/>
  </div>
</template>

<script>
import ApolloHeader from './components/ApolloHeader'
import CollectionList from './components/CollectionList'
import ApolloFooter from './components/ApolloFooter'

export default {
  components: {
    'apollo-header': ApolloHeader,
    'collection-list': CollectionList,
    'apollo-footer': ApolloFooter
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
