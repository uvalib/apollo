<template>
  <div id="app">
    <Header/>
    <Collections/>
    <Footer/>
  </div>
</template>

<script>
import Header from './components/Header'
import Collections from './components/Collections'
import Footer from './components/Footer'

export default {
  name: 'App',
  components: {
    Header,
    Collections,
    Footer
  }
}
</script>

<style>
body {
  background-color: #002F6C;
  padding: 0;
  margin: 0;
}
#app {
  font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
