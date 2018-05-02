<template>
  <div class="page-header">
    <a v-show="back" class="back" href="/">&larr; Back</a>
    <span class="user">FOO</span>
    <h2 class="page-header">{{ main }}</h2>
    <p class="page-header"><span v-html="sub"></span></p>
  </div>
</template>

<script>
  export default {
    name: 'page-header',
    props: {
      main: String,
      sub: String,
      back: Boolean }
  }
</script>

<style scoped>
  div.page-header {
    border-bottom: 1px solid rgb(229, 114, 0);
    margin-bottom: 30px;
    padding-bottom: 10px;
    position: relative;
  }
  span.user {
    position: absolute;
    right: 0px;
    top: -10px;
  }
  h2.page-header {
    color: rgb(229, 114, 0);
    margin: 0 0 5px 0;
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
