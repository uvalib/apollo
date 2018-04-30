<template>
  <div class="detail-wrapper">
    <apollo-error v-if="error" :message="errorMsg"></apollo-error>
    <template v-else>
      <div class="page-header">
        <h2 class="page-header">Collection</h2>
        <p class="page-header">{{ id }}: <b>{{ title }}</b></p>
      </div>
      <loading-spinner v-if="loading"/>
      <div v-else class="content">
      </div>
    </template>
  </div>
</template>

<script>
  import axios from 'axios'
  import LoadingSpinner from './LoadingSpinner'
  import ApolloError from './ApolloError'

  export default {
    name: 'collection-detail',
    components: {
      'loading-spinner': LoadingSpinner,
      'apollo-error': ApolloError
    },
    props: {
      id: String,
      title: String
    },
    data: function () {
      return {
        collection: {},
        loading: false,
        error: null,
      }
    },
    created: function () {
      this.loading = true;
      var self = this;
      axios.get("/api/collections/"+this.id).then((response)  =>  {
        this.loading = false;
        this.collection = response.data
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
  div.detail-wrapper {
    background-color: white;
    padding: 20px;
  }
</style>
