<template>
  <div class="loading">
    <img src="../assets/spinner2.gif"/>
    <span>Loading...</span>
  </div>
</template>

<script>
  export default {
    name: 'loading-spinner'
  }
</script>

<style scoped>
  .loading {
    padding: 25px 50px;
  }
  img {
    vertical-align: middle;
    padding-right: 20px;
  }
</style>
