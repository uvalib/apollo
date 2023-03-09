<template>
  <ApolloHeader/>
  <router-view></router-view>
</template>

<script setup>
import ApolloHeader from './components/ApolloHeader.vue'
</script>

<style>
body {
  background-color: white;
  padding: 0;
  margin: 0;
}
</style>
