<template>
  <div class="collections">
    <div class="page-header">
      <h2>Collections</h2>
      <p>The following are all of the digitized serials managed by <span class="apollo">Apollo</span>:</p>
    </div>
    <div v-if="loading">
      Loading.....
    </div>
    <table v-else class="collection-list">
      <tr><th class="right">PID</td><th>Title</th></tr>
      <tr v-for="item in collections">
        <td class="right">{{ item.pid }}</td><td>{{ item.title }}</td>
      </tr>
    </table>
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
      axios.get("/api/collections").then((response)  =>  {
        this.loading = false;
        this.collections = response.data;
      }, (error)  =>  {
        this.loading = false;
        console.log(error);
      })
    }
  }
</script>

<style scoped>
  .page-header {
    border-bottom: 1px solid rgb(229, 114, 0);
  }
  h2 {
    color: rgb(229, 114, 0);
    margin: 0 0 5px 0;
  }
  p {
    margin: 0 0 10px 15px;
  }
  span.apollo {
    font-family: 'Righteous', cursive;
    color: #2c3e50;
  }
  div.collections {
    background: white;
    font-family: 'Avenir', Helvetica, Arial, sans-serif;
    color: #2c3e50;
    padding: 20px;
  }
  table {
    margin: 15px;
    border-collapse: collapse;
  }
  td.right, th.right  {
    text-align: right;
    padding-right: 10px;
    border-right: 1px solid #ccc;
  }
  td {
    padding: 5px 10px 2px 10px;
  }
  th {
    background-color: #f5f5f5;
    padding: 5px 5px 2px 10px;
    text-align: left;
    border-bottom: 1px solid #ccc;
  }
</style>
