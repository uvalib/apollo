<template>
  <div class="page-header">
    <router-link v-show="back" class="back" :to="{ name: 'home'}">&larr; Back</router-link>
    <span class="user">{{ this.user }}</span>
    <h2 class="page-header">{{ main }}</h2>
    <p class="page-header"><span v-html="sub"></span></p>
  </div>
</template>

<script>
  export default {
    props: {
      main: String,
      sub: String,
      back: Boolean
    },
    data: function () {
      return {
        user: localStorage.getItem("user")
      }
    }
  }
</script>

<style scoped>
  div.page-header {
    border-bottom: 1px solid rgb(229, 114, 0);
    margin-bottom: 15px;
    padding-bottom: 5px;
    position: relative;
  }
  span.user {
    position: absolute;
    right: 0px;
    top: -10px;
  }
  h2.page-header {
    color: rgb(229, 114, 0);
    margin: 0;
  }
  p.page-header {
    margin: 0 0 0 30px;
    color: #666;
  }
  a.back {
    color: #2c3e50;;
    text-decoration: none;
    font-size: 0.9em;
    position: relative;
    top: -10px;
  }
  a.back:hover {
    text-decoration: underline;
  }
</style>
