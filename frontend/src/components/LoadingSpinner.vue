<template>
  <div class="loading">
    <img src="../assets/spinner2.gif"/>
    <span>{{message}}...</span>
  </div>
</template>

<script setup>
const props = defineProps({
   message: {
      type: String,
      required: true
   }
})
</script>

<style scoped>
  .loading {
    padding: 25px 50px;
    background: white;
  }
  img {
    vertical-align: middle;
    padding-right: 20px;
  }
</style>
