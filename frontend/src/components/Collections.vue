<template>
  <div class="collections">
    <h1>TODO</h1>
  </div>
</template>

<script>
  import axios from 'axios';

  export default {
    name: 'Collections',
    data: function () {
      return {
        collections: [],
        loading: false
      }
    },
    created: function () {
      this.loading = true;
      // alert("Get...")
      axios.get("/api/collections").then((response)  =>  {
        this.loading = false;
        console.log(response);
      }, (error)  =>  {
        this.loading = false;
        console.log(error);
      })
    }
  }
</script>

<style scoped>
  div.collections {
    background: white;
    font-family: 'Avenir', Helvetica, Arial, sans-serif;
    color: #2c3e50;
    padding: 20px;
  }
</style>
