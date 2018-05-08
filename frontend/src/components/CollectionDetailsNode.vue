<template>
  <li>
    <span v-if="isFolder" class="icon" @click="toggle" :class="{ plus: open==false, minus: open==true}"></span>
    <table class="node">
      <tr class="attribute">
        <td class="label">Type:</td><td class="data">{{ model.name.value }}</td>
      </tr>
      <tr v-for="(attribute, index) in model.attributes"  class="attribute">
        <template v-if="attribute.name.value !='digitalObject'">
          <td class="label">{{ attribute.name.value }}:</td><td class="data">{{ attribute.value }}</td>
        </template>
        <template v-else>
          <td class="label"></td>
          <td :data-uri="attribute.value" @click="digitalObjectClicked" class="data dobj">View Digitial Object</td>
        </template>
      </tr>
    </table>
    <ul v-if="open" v-show="open">
      <template v-for="(child, index) in model.children">
        <details-node :model="child" :depth="depth+1"></details-node>
      </template>
    </ul>
  </li>
</template>

<script>
  import axios from 'axios'

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
    mounted() {
      window.addEventListener("scroll", this.handleScroll);
    },
    destroyed() {
      window.removeEventListener("scroll", this.handleScroll);
    },
    methods: {
      handleScroll: function(event) {
        var viewer = $('#object-viewer')
        let origVal = viewer.data("origTop")
        // console.log("OV: ["+origVal+"]")
        if ( !origVal ) {
          let ot = $('#object-viewer').offset().top
          console.log("set origTop to: "+ot)
          viewer.data("origTop", ot)
        }

        let scrollTop= $(window).scrollTop();
        // console.log("SCROLL TOP: "+scrollTop);

        // var isPositionFixed = ($el.css('position') == 'fixed');
        if ( scrollTop >= 252 ) {
           let xxx = $('#object-viewer').offset().top
           // console.log("VIEW TOP" +xxx)
           viewer.offset({top: scrollTop+15});
        } else {
           viewer.offset({top: viewer.data("origTop")});
        }
        // TODO maybe? fix scroll off screen bottom. Don't think needed tho
      },
      toggle: function () {
        if (this.isFolder) {
          this.open = !this.open
        }
      },
      digitalObjectClicked: function(event) {
        // make sure only one node is marked as selected
        $(".selected").removeClass("selected")
        let node = $(event.target).closest(".node")
        node.addClass("selected")

        // grab the oembedURI and request embedding info
        let oembedUri = event.target.getAttribute('data-uri')
        axios.get(oembedUri).then((response)  =>  {
          let dv = $("#object-viewer")
          dv.empty();
          // set a global flag to make the browser think JS jas not all been
          // loaded. Without this, the JS file included in the response
          // will not load, and the viewer will not render
          window.embedScriptIncluded = false;
          dv.append( $( response.data.html) )
        }).catch((error) => {
          if ( error.message ) {
            alert(error.message)
          } else {
            alert(error.response.data)
          }
        })
      }
    }
  }
</script>

<style scoped>

  table.node {
    padding: 5px 0 5px 5px;
    margin: 0;
    display: block;
    overflow-wrap: break-word;
    word-wrap: break-word;
    hyphens: auto;
    border-bottom: 1px solid #ccc;
    border-left: 1px solid #ccc;
  }
  table.node td {
    padding: 5px 10px 5px 0;
  }
  table.node.selected {
    border-left: 2px solid #5c5;
    border-bottom: 2px solid #5c5;
    background: #f5fff5;
  }
  .dobj {
    cursor: pointer;
    color: #999;
    font-weight: bold;
    font-style: italic;
  }
  .dobj:hover {
    text-decoration: underline;
  }
  table.node td.label {
    text-transform: capitalize;
    margin-right: 15px;
    font-weight: bold;
    text-align: right;
    padding: 5px 10px 5px 10px;
  }
  table.node td.data {
    text-transform: capitalize;
  }
  div.content ul, div.content li {
    font-size: 14px;
    list-style-type: none;
    position: relative;
  }
  span.icon {
    display: inline-block;
    width: 18px;
    height: 18px;
    position: absolute;
    left: -35px;
    top: 10px;
    /* border: 1px solid #ccc; */
    padding: 4px 11px 4px 10px;
    /* border-right: 1px solid white; */
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
