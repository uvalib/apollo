<template>
  <li>
    <div class="node">
      <div class="attribute">
        <label>Type:</label><span class="data">{{ model.name.value }}</span>
      </div>
      <div v-for="(attribute, index) in model.attributes">
        <label>{{ attribute.name.value }}:</label><span class="data">{{ attribute.value }}</span>
      </div>
    </div>
    <ul>
      <template v-for="(child, index) in model.children">
        <details-node :model="child" :depth="0"></details-node>
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
    }
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
  }
  div.node {
    padding: 10px;
    border:1px solid #ddd;
    margin: 15px;
  }
</style>
