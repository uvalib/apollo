<template>
  <ApolloError v-if="errorMsg" :message="errorMsg"/>
  <div v-else class="collections">
    <PageHeader
      main="Collections"
      sub="The following are all of the digitized serials managed by Apollo"
    />
    <LoadingSpinner v-if="loading" message="Loading collections"/>
    <div v-else class="content">
      <table class="collection-list">
        <tr><th></th><th class="right">PID</th><th>Title</th></tr>
        <tr v-for="item in collections" :key="item.pid">
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
  </div>
</template>

<script>
  import axios from 'axios'
  import PageHeader from '@/components/PageHeader'
  import LoadingSpinner from '@/components/LoadingSpinner'

  export default {
    components: {
      PageHeader,
      LoadingSpinner
    },
    data: function () {
      return {
        collections: [],
        loading: true,
        errorMsg: null
      }
    },
    created: function () {
      axios.get("/api/collections").then((response)  =>  {
        this.collections = response.data
      }).catch((error) => {
        this.errorMsg = error.response.data
      }).finally(() => {
        this.loading = false
      })
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
