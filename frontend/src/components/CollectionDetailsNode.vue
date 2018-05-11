<template>
  <li>
    <span v-if="isFolder" class="icon" @click="toggle" :class="{ plus: open==false, minus: open==true}"></span>
    <table class="node">
      <tr class="attribute">
        <td class="label">Type:</td>
        <td class="data">
          {{ model.name.value }}
          <div class="formats">
            <template v-if="jsonLink()">
              <a class="raw" :href="jsonLink()" target="_blank">JSON</a>
            </template>
            <template v-if="sirsiLink()">
              <a class="sirsi" :href="sirsiLink()" target="_blank">Sirsi</a>
            </template>
          </div>
        </td>
      </tr>
      <tr v-for="(attribute, index) in model.attributes"  class="attribute">
        <template v-if="attribute.name.value !='digitalObject'">
          <template v-if="attribute.valueURI">
            <td class="label">{{ attribute.name.value }}:</td>
            <td class="data">
              <a :href="attribute.valueURI" class="uri" target="_blank">{{ attribute.value }}&nbsp;<i class="fas fa-external-link-alt"></i></a>
            </td>
          </template>
          <template v-else>
            <td class="label">{{ attribute.name.value }}:</td><td class="data">{{ attribute.value }}</td>
          </template>
        </template>
        <template v-else>
          <td class="label"></td>
          <td :data-uri="attribute.value" @click="digitalObjectClicked" class="pure-button pure-button-primary data dobj">View Digitial Object</td>
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
  import EventBus from './EventBus'

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
      componentPID: function() {
        for (var idx in this.model.attributes) {
          var attr = this.model.attributes[idx]
          if (attr.name.value === "componentPID") {
            return attr.value
          }
        }
        return ""
      },

      jsonLink: function() {
        // This should only return a URL for nodes that
        // are top level. A top level node will have a barcode and/or key
        for (var idx in this.model.attributes) {
          let attr = this.model.attributes[idx]
          if (attr.name.value === "barcode" || attr.name.value === "catalogKey") {
            return "/api/collections/"+this.model.pid
          }
        }
        return ""
      },

      sirsiLink: function(model) {
        // This should only return a URL for nodes that
        // are top level. A top level node will have a barcode and/or key
        let barcode=""
        let catalogKey = ""
        for (var idx in this.model.attributes) {
          let attr = this.model.attributes[idx]
          if (attr.name.value === "barcode"){
            barcode = attr.value
          }
          if (attr.name.value === "catalogKey") {
            catalogKey = attr.value
          }
        }

        if (barcode.length > 0) {
          return "http://solr.lib.virginia.edu:8082/solr/core/select/?q=barcode_facet:"+barcode
        }
        if (catalogKey.length > 0) {
          return "http://solr.lib.virginia.edu:8082/solr/core/select/?q=id:"+catalogKey
        }
        return ""
      },

      handleScroll: function(event) {
        // Keep the viewer on screen as the user scrolls through
        // the (potentially) long list of nodes in the collection.
        var viewer = $('#viewer-wrapper')
        let origVal = viewer.data("origTop")
        // console.log("OV: ["+origVal+"]")
        if ( !origVal ) {
          let ot = $('#viewer-wrapper').offset().top
          viewer.data("origTop", ot)
        }

        let scrollTop= $(window).scrollTop();
        // console.log("SCROLL TOP: "+scrollTop);

        // var isPositionFixed = ($el.css('position') == 'fixed');
        if ( scrollTop >= 252 ) {
           let xxx = $('#viewer-wrapper').offset().top
           // console.log("VIEW TOP" +xxx)
           viewer.offset({top: scrollTop+15});
        } else {
           viewer.offset({top: viewer.data("origTop")});
        }
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
        EventBus.$emit('viewer-clicked')

        // grab the oembedURI and request embedding info
        let oembedUri = event.target.getAttribute('data-uri')
        axios.get(oembedUri).then((response)  =>  {
          let dv = $("#object-viewer")
          dv.empty()
          // set a global flag to make the browser think JS jas not all been
          // loaded. Without this, the JS file included in the response
          // will not load, and the viewer will not render
          window.embedScriptIncluded = false
          dv.append( $( response.data.html) )
          EventBus.$emit('viewer-opened', this.componentPID() )
        }).catch((error) => {
          if ( error.message ) {
            EventBus.$emit('viewer-error', error.message)
          } else {
            EventBus.$emit('viewer-error', error.response.data)
          }
        })
      }
    }
  }
</script>

<style scoped>
  div.formats {
    position: absolute;
    right: 3px;
    top: 0;
  }
  .sirsi, .raw {
    padding: 2px 8px;
    border-radius: 10px;
    background: #0078e7;
    color: white;
    opacity: 0.6;
    cursor: pointer;
    text-decoration: none;
  }
  .sirsi:hover, .raw:hover {
    opacity: 1;
  }
  .raw {
    background: rgb(240, 130, 40);
    opacity: 0.8;
  }
  i.fa-external-link-alt {
    margin-left:5px;
    color: #999;
    opacity: 0.5;
  }
  a.uri {
    color: cornflowerblue;
    text-decoration: none;
    font-weight: bold;
  }
  a.uri:hover {
    text-decoration: underline;
  }
  table.node {
    padding: 5px 0 5px 5px;
    margin: 0;
    display: table;
    width: 100%;
    overflow-wrap: break-word;
    word-wrap: break-word;
    hyphens: auto;
    border-bottom: 1px solid #ccc;
    border-left: 1px solid #ccc;
    font-size: 0.9em;
  }
  table.node td {
    padding: 5px 10px 5px 0;
  }
  table.node tr {
    padding: 5px 10px 5px 0;
  }
  table.node.selected {
    border-left: 2px solid #7af;
    border-bottom: 2px solid #7af;
    background: #f0f7ff;
  }
  td.sirsi-link {
    text-align: right;
  }
  td.dobj.pure-button {
    padding: 4px 10px;
    margin: 4px 0;
    opacity: 0.6;
    border-radius: 10px;
  }
  td.dobj.pure-button:hover {
    opacity: 1;
  }
  table.node td.label {
    text-transform: capitalize;
    margin-right: 15px;
    font-weight: bold;
    text-align: right;
    padding: 5px 10px 5px 10px;
    width: 35%;
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
    padding: 4px 11px 4px 10px;
    z-index: 100;
    cursor: pointer;
    opacity: 0.4;
  }
  span.icon:hover {
    opacity: 0.7;
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
