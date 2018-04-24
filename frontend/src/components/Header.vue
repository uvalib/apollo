<template>
   <div class="header">
      <h1>Apollo</h1>
      <span class="subtitle">Digitized Serials Management System</span>
      <a target="_blank" href="https://library.virginia.edu"><img src="../assets/uvalogo.png"/></a>
   </div>
</template>

<script>
export default {
  name: 'Header'
}
</script>

<style scoped>
div.header {
   background-color: #002F6C;
   color: white;
   padding:20px;
   border-bottom: 1px solid #d2e5f7;
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
