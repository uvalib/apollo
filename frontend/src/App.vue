<template>
   <div id="app">
      <Header/>
      <div class="main">
         <h1>Work In Progress</h1>
      </div>
      <Footer/>
   </div>
</template>

<script>
import Header from './components/Header'
import Footer from './components/Footer'

export default {
   name: 'App',
   components: {
      Header,
      Footer
   }
}
</script>

<style>
body {
   /* background-color: #d2e5f7; */
   background-color: #002F6C;
   padding: 0;
   margin: 0;
}
div.main {
   background: white;
   font-family: 'Avenir', Helvetica, Arial, sans-serif;
   color: #2c3e50;
   padding: 20px;   
}
#app {
   font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
   -webkit-font-smoothing: antialiased;
   -moz-osx-font-smoothing: grayscale;
}
</style>
