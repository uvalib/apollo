<template>
  <div class="hit">
    <div class="collection"><a :href="getHitURL">{{hit.collection_pid}} - {{hit.title}}</a></div>
    <div class="match"><b>{{hit.item_type}}:</b> {{hit.match}}</div>
  </div>
</template>

<script>
  export default {
      name: "ApolloSearchHit",
      props: {
         hit: Object,
      },
      computed: {
        getHitURL: function() {
          let idx = this.hit.item_url.indexOf("/collections")
          if (idx > 0) {
            return this.hit.item_url.substring(idx, this.hit.item_url.length)
          }
          return this.hit.item_url
        }
      }
   }
</script>

<style scoped>
  div.hit {
    font-size: 0.85em;
    padding-bottom: 5px;
    color: #666;
  }
  a {
    color: cornflowerblue;
    text-decoration: none;
    font-weight: bold;
  }
  a:hover {
    text-decoration: underline;
  }
</style>
