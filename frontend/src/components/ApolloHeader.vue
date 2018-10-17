<template>
   <div class="header">
      <router-link :to="{ name: 'home'}"><h1>Apollo</h1></router-link>
      <span class="subtitle">Digitized Serials Management System</span>
      <a target="_blank" href="https://library.virginia.edu"><img src="../assets/uvalogo.png"/></a>
   </div>
</template>

<script>
export default {
}
</script>

<style scoped>
div.header {
   background-color: #002F6C;
   color: white;
   padding:20px;
   border-bottom: 1px solid #d2e5f7;
}
a {
  color: white;
  text-decoration: none;
}
h1 {
   margin: 0;
   font-family: 'Righteous', cursive;
   font-size: 40px;
}
img {
  position: absolute;
  top: 15px;
  right: 15px;
  max-height: 75px;
}
.subtitle {
   display: inline-block;
   margin: 5px 0 0 15px;
}
</style>
