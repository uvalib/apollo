<template>
  <li>
    <span v-if="isFolder" class="icon" @click="toggle" :class="{ plus: open==false, minus: open==true}"></span>
    <div class="node">
      <div class="attribute">
        <label>Type:</label><span class="data">{{ model.name.value }}</span>
      </div>
      <div v-for="(attribute, index) in model.attributes"  class="attribute">
        <template v-if="attribute.name.value !='digitalObject'">
          <label>{{ attribute.name.value }}:</label><span class="data">{{ attribute.value }}</span>
        </template>
        <span v-else :data-uri="attribute.value" @click="digitalObjectClicked" class="dobj">Digitial Object</span>
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
    methods: {
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
        // var hack = '<div class="uv" data-uri="https://tracksys.lib.virginia.edu:8080/uva-lib:2528443" data-canvasindex="0" style="width:800px; height:600px;"></div>'
        // var dv = $("#object-viewer")
        // dv.empty()
        // dv.append( $(hack) )
        // var script = document.createElement("script");
        // script.type = 'text/javascript'
        // script.src = "https://doviewer.lib.virginia.edu/web/viewer/lib/embed.js"
        // script.id = "embedUV"
        // window.embedScriptIncluded = false;
        // dv.append( $(script) )


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
    },
  }
</script>

<style scoped>
  span.dobj {
    cursor: pointer;
  }
  span.dobj:hover {
    text-decoration: underline;
  }
  label {
    text-transform: capitalize;
    margin-right: 15px;
    font-weight: bold;
  }
  span.data {
    text-transform: capitalize;
  }
  div.content ul, div.content li {
    font-size: 14px;
    list-style-type: none;
    position: relative;
  }
  div.node {
    padding: 10px 0 10px 10px;;
    margin: 0;
    position: relative;
    overflow-wrap: break-word;
    word-wrap: break-word;
    hyphens: auto;
    border-bottom: 1px solid #ccc;
    border-left: 1px solid #ccc;
    line-height: 1.75em;
  }
  div.node.selected {
    border-left: 2px solid #ccf;
    border-bottom: 2px solid #ccf;
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
