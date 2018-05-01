<template>
  <li>
    <span v-if="isFolder" class="icon" @click="toggle" :class="{ plus: open==false, minus: open==true}"></span>
    <div class="node">
      <div class="attribute">
        <label>Type:</label><span class="data">{{ model.name.value }}</span>
      </div>
      <div v-for="(attribute, index) in model.attributes">
        <label>{{ attribute.name.value }}:</label><span class="data">{{ attribute.value }}</span>
      </div>
    </div>
    <ul v-if="open" v-show="open">
      <template v-for="(child, index) in model.children">
        <details-node :model="child" :depth="depth+1"></details-node>
      </template>
    </ul>
  </li>
</template>

<script>
  export default {
    name: 'details-node',
    props: {
      model: Object,
      depth: Number
    },
    data: function () {
      return {
        open: this.depth < 1
      }
    },
    computed: {
      isFolder: function () {
        return this.model.children && this.model.children.length
      }
    },
    methods: {
      toggle: function () {
        if (this.isFolder) {
          this.open = !this.open
        }
      }
    },
  }
</script>

<style scoped>
  label {
    text-transform: capitalize;
    margin-right: 15px;
    font-weight: bold;
  }
  span.data {
    text-transform: capitalize;
  }
  ul, li {
    list-style-type: none;
    position: relative;
  }
  div.node {
    padding: 10px;
    border:1px solid #ddd;
    margin: -1px 1px;
    position: relative;
  }
  span.icon {
    display: inline-block;
    width: 18px;
    height: 18px;
    position: absolute;
    left: -39px;
    top: 0px;
    border: 1px solid #ccc;
    padding: 4px 11px 4px 10px;
    /* border-radius: 15px 0 0 15px; */
    border-right: 1px solid white;
    z-index: 100;
    cursor: pointer;
  }
  span.icon.plus {
    background: url(../assets/plus.png);
    background-repeat: no-repeat;
    background-position: 5px,6px;
  }
  span.icon.minus {
    background: url(../assets/minus.png);
    background-repeat: no-repeat;
    background-position: 5px,6px;
  }
</style>
