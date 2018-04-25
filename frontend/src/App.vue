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
  /* background-color: #e0e0e0; */
  padding: 0;
  margin: 0;
}
#app {
  font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  /* max-width: 1024px;
  margin: 0 auto;
  border-right: 1px solid #999;
  border-left: 1px solid #999; */
}
</style>
