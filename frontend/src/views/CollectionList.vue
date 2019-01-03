<template>
  <ApolloError v-if="hasError" :message="error"/>
  <div v-else class="collections">
    <PageHeader
      main="Collections"
      sub="The following are all of the digitized serials managed by Apollo"
    />
    <LoadingSpinner v-if="isLoading" message="Loading collections"/>
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
  import ApolloError from '@/views/ApolloError'
  import PageHeader from '@/components/PageHeader'
  import LoadingSpinner from '@/components/LoadingSpinner'
  import { mapGetters } from 'vuex'

  export default {
    components: {
      ApolloError,
      PageHeader,
      LoadingSpinner
    },
    computed: { 
      // map getters for stuff like this.$store.blah to simple named properties
      // not strictly necessary, but cleaner syntax/less code. Just list getter name
      // to make the mapping. If you want to change the name do it something like
      // doneCount: 'doneTodosCount'
      ...mapGetters([
        'isLoading',
        'collections',
        'hasError',
        'error'
      ])
    },
    created: function () {
      this.$store.dispatch('getCollections')
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
