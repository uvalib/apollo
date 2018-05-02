<template>
  <div class="collections">
    <apollo-error v-if="error" :message="errorMsg"></apollo-error>
    <template v-else>
      <page-header
        main="Collections"
        sub="The following are all of the digitized serials managed by Apollo:"
        :back="false"></page-header>
      <loading-spinner v-if="loading"/>
      <div v-else class="content">
        <table class="collection-list">
          <tr><th></th><th class="right">PID</td><th>Title</th></tr>
          <tr v-for="item in collections">
            <td class="icon">
              <router-link :to="{ name: 'collections', params: {id: item.pid, title: item.title}}">
                <img class="detail" src="../assets/detail.png"/>
              </router-link>
            </td>
            <td class="right">{{ item.pid }}</td>
            <td>{{ item.title }}</td>
          </tr>
        </table>
      </div>
    </template>
  </div>
</template>

<script>
  import axios from 'axios'
  import LoadingSpinner from './LoadingSpinner'
  import ApolloError from './ApolloError'
  import PageHeader from './PageHeader'

  export default {
    name: 'collection-list',
    components: {
      'loading-spinner': LoadingSpinner,
      'apollo-error': ApolloError,
      'page-header': PageHeader
    },
    data: function () {
      return {
        collections: [],
        loading: false,
        error: null,
      }
    },
    created: function () {
      this.loading = true;
      var self = this;
      axios.get("/api/collections").then((response)  =>  {
        this.loading = false;
        this.collections = response.data
        self.error = null
      }).catch(function (error) {
        self.loading = false;
        self.error = true;
        self.errorMsg = error.response.data;
      });
    }
  }
</script>

<style scoped>
  div.collections {
    background: white;
    padding: 20px;
  }
  table {
    margin: 15px;
    border-collapse: collapse;
    border-right: 1px solid #ddd;
    border-left: 1px solid #ddd;
  }
  td.right, th.right  {
    text-align: right;
    padding-right: 10px;
    border-right: 1px solid #ccc;
  }
  td.icon {
    padding: 0 10px;
  }
  img.detail {
    vertical-align: middle;
    opacity: 0.6;
    cursor: pointer;
  }
  img.detail:hover {
    opacity: 1;
  }
  td {
    cursor: default;
    padding: 10px 10px 8px 10px;
    border-bottom: 1px solid #ddd;
  }
  th {
    background-color: #f5f5f5;
    padding: 10px 10px 8px 10px;
    text-align: left;
    border-bottom: 1px solid #ccc;
    border-top: 1px solid #ddd;
  }
</style>
