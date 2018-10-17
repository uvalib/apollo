<template>
  <li :id="model.pid">
    <span v-if="isFolder" class="icon" @click="toggle" :class="{ plus: open==false, minus: open==true}"></span>
    <table class="node">
      <tr class="attribute">
        <td class="label">PID:</td>
        <td class="data">{{ model.pid }}</td>
      </tr>
      <tr class="attribute">
        <td class="label">Type:</td>
        <td class="data">
          {{ model.type.name }}
        </td>
      </tr>
      <tr v-for="attribute in model.attributes" :key="attribute.pid" class="attribute">
        <template v-if="attribute.type.name !='digitalObject'">
          <td class="label">{{ attribute.type.name }}:</td>
          <td class="data">
            <span v-html="renderAttributeValue(attribute)"></span>
            <span v-if="showMore(attribute)" class='show-more' @click="moreClicked" >more</span>
          </td>
        </template>
        <template v-else>
          <td class="label"></td>
          <td :data-uri="getDoViewerURL(attribute)" @click="digitalObjectClicked" class="pure-button pure-button-primary data dobj">View Digitial Object</td>
        </template>
      </tr>
    </table>
    <ul v-if="open" v-show="open">
      <template v-for="child in model.children">
        <details-node :key="child.pid" :model="child" :depth="depth+1"></details-node>
      </template>
    </ul>
  </li>
</template>

<script>
  import axios from 'axios'
  import EventBus from './EventBus'

  export default {
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
      EventBus.$on("expand-node", this.handleExpandNodeEvent)
      window.addEventListener("scroll", this.handleScroll)
      EventBus.$emit('node-mounted', this.model.pid)
    },

    destroyed() {
      window.removeEventListener("scroll", this.handleScroll)
    },

    methods: {
      handleExpandNodeEvent: function(pid) {
        if ( this.model.pid == pid) {
          if ( this.isFolder && this.open == false ) {
            this.toggle()
          }
        }
      },
      getDoViewerURL: function(attribute) {
        if (attribute.values[0].value.includes("https://")) {
          return attribute.values[0].value
        }
        // conert JSON to something like this:
        // https://doviewer.lib.virginia.edu/oembed?url=https%3A%2F%2Fdoviewer.lib.virginia.edu%2Fimages%2Fuva-lib%3A2528443
        let json = JSON.parse(attribute.values[0].value)
        let qp = encodeURIComponent(process.env.VUE_APP_DOVIEWER_URL+"/"+json.type+"/"+json.id)
        let url = process.env.VUE_APP_DOVIEWER_URL+"/oembed?url="+qp
        return url
      },
      showMore: function(attribute) {
        if (attribute.values.length > 1) return false
        return attribute.values[0].value.length > 150
      },
      moreClicked: function(event) {
        let btn = $(event.currentTarget)
        if (btn.text() == "more") {
          let parent = $(event.currentTarget).closest("td")
          let txt = parent.find(".long-val")
          txt.text(txt.data("full"))
          btn.text("less")
        } else {
          let parent = $(event.currentTarget).closest("td")
          let txt = parent.find(".long-val")
          txt.text(txt.data("full").substring(0,150)+"...")
          btn.text("more")
        }
      },
      renderAttributeValue: function(attribute) {
        let out = ""
        for (var idx in attribute.values) {
          let val = attribute.values[idx]
          if (out.length > 0) out += "<span>, </span>"
          if (val.valueURI) {
            out += "<a class='uri' href='"+val.valueURI+"' target='_blank'>"+val.value+"</a>"
          } else {
            if (val.value.length < 150) {
              out += "<span>"+val.value+"</span>"
            } else {
              out += "<span class='long-val' data-full='"+val.value+"'>"+val.value.substring(0,150)
              out += "...</span>"
            }
          }
        }
        return out
      },

      externalPID: function() {
        for (var idx in this.model.attributes) {
          var attr = this.model.attributes[idx]
          if (attr.type.name === "externalPID") {
            return attr.values[0].value
          }
        }
        return ""
      },

      handleScroll: function() {
        // Keep the viewer on screen as the user scrolls through
        // the (potentially) long list of nodes in the collection.
        var viewer = $('#viewer-wrapper')
        if (viewer.length === 0 ) return

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
           // console.log("VIEW TOP" +$('#viewer-wrapper').offset().top)
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
          EventBus.$emit('viewer-opened', this.externalPID() )
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
    min-width:90px;
  }
  div.content ul, div.content li {
    font-size: 14px;
    list-style-type: none;
    position: relative;
  }
  span.icon {
    width: 18px;
    height: 18px;
    position: absolute;
    left: 8px;
    top: 8px;
    padding: 0;
    z-index: 100;
    cursor: pointer;
    opacity: 0.4;
    background-repeat: no-repeat;
    background-position: 0;
  }
  span.icon:hover {
    opacity: 0.7;
  }
  span.icon.plus {
    background: url(../assets/plus.png);
  }
  span.icon.minus {
    background: url(../assets/minus.png);
  }
  td.data {
    position: relative;
  }
  td.dobj {
    float:right;
  }
</style>
