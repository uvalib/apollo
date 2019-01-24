<template>
  <div class="hit">
    <div><a :href="hit.item_url">{{hit.pid}}</a> - {{hitTitle}}</div>
    <p v-html="hitSnippet"></p>
  </div>
</template>

<script>
  import { mapGetters } from 'vuex'
  export default {
      name: "ApolloSearchHit",
      props: {
         hit: Object,
      },
      computed: {
        ...mapGetters([
            'searchQuery',
        ]),
        hitTitle() {
          if (this.hit.title) {
            return this.hit.title
          } else {
            return this.hit.match
          }
        },
        hitSnippet() {
          let matchStr = this.hit.match
          let lcMatch = this.hit.match.toLowerCase()
          let p0 = lcMatch.indexOf(this.searchQuery)
          let p1 = p0+this.searchQuery.length
          let tStyle="text-transform:capitalize;font-weight:600"
          let hStyle="color:rgb(229, 114, 0);font-weight:500;"
          let out = `<span style="${tStyle}">${this.hit.match_type}: </span>${matchStr.substring(0,p0)}<b style="${hStyle}">${matchStr.substring(p0,p1)}</b>${matchStr.substring(p1)}`
          return out
         }
      }
   }
</script>

<style scoped>
p {
  margin: 3px 0 15px 0;
}
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
