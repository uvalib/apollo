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
          <td colspan="2" class="do-buttons">
            <a v-if="hasIIIFManifest" class="do-button" :href="iiifManufestURL" target="_blank">IIIF Manifest</a>
            <span :data-uri="getCurioURL(attribute)"
              @click="digitalObjectClicked"
              class="do-button">View Digitial Object</span>
          </td>
        </template>
      </tr>
    </table>
    <ul v-if="open" v-show="open">
      <CollectionDetailsItem  v-for="child in model.children" :key="child.pid" :model="child" :depth="depth+1"/>
    </ul>
  </li>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
   model: {
      type: Object,
      required: true
   },
   depth: {
      type: Number,
      required: true
   },
})

const open = computed( () => {
   return props.depth < 1
})
const isFolder = computed( () => {
   return props.model.children && props.model.children.length
})
const iiifManufestURL = computed( () => {
   return process.env.VUE_APP_IIIF_MAN_URL+"/"+externalPID()+"/manifest.json"
})
const hasIIIFManifest = computed( () => {
   if (props.model.type.name === "item") {
      return false
   }
   return true
})

  //   mounted() {
  //     EventBus.$on("expand-node", this.handleExpandNodeEvent)
  //     EventBus.$on('collapse-all', this.handleCollapseAll)
  //     EventBus.$emit('node-mounted', props.model.pid)
  //   },

  //   destroyed() {
  //     EventBus.$emit('node-destroyed')
  //   },

  //   methods: {
  //     handleCollapseAll: function() {
  //       if ( this.isFolder && this.open ) {
  //         this.toggle()
  //       }
  //     },
  //     handleExpandNodeEvent: function(pid) {
  //       if ( props.model.pid == pid) {
  //         if ( this.isFolder && this.open === false ) {
  //           this.toggle()
  //         }
  //       }
  //     },
  //     getCurioURL: function(attribute) {
  //       if (attribute.values[0].value.includes("https://")) {
  //         return attribute.values[0].value
  //       }
  //       // conert JSON to something like this:
  //       // https://curio.lib.virginia.edu/oembed?url=https%3A%2F%2Fcurio.lib.virginia.edu%2Fview%2Fuva-lib%3A2528443
  //       let json = JSON.parse(attribute.values[0].value)
  //       let qp = encodeURIComponent(process.env.VUE_APP_CURIO_URL+"/view/"+json.id)
  //       let url = process.env.VUE_APP_CURIO_URL+"/oembed?url="+qp
  //       return url
  //     },
  const showMore = ((attribute) => {
    if (attribute.values.length > 1) return false
    return attribute.values[0].value.length > 150
  })

  //     moreClicked: function(event) {
  //       let btn = $(event.currentTarget)
  //       if (btn.text() == "more") {
  //         let parent = $(event.currentTarget).closest("td")
  //         let txt = parent.find(".long-val")
  //         txt.text(txt.data("full"))
  //         btn.text("less")
  //       } else {
  //         let parent = $(event.currentTarget).closest("td")
  //         let txt = parent.find(".long-val")
  //         txt.text(txt.data("full").substring(0,150)+"...")
  //         btn.text("more")
  //       }
  //     },
  const renderAttributeValue = ((attribute) => {
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
  })

  const externalPID = (() => {
    for (var idx in props.model.attributes) {
      var attr = props.model.attributes[idx]
      if (attr.type.name === "externalPID") {
        return attr.values[0].value
      }
    }
    return ""
  })

// const toggle =  (()=> {
//    if (isFolder.value) {
//       this.open = !this.open
//    }
// })

  //     digitalObjectClicked: function(event) {
  //       // make sure only one node is marked as selected
  //       $(".selected").removeClass("selected")
  //       let node = $(event.target).closest(".node")
  //       node.addClass("selected")
  //       this.$store.commit("setViewerLoading", true)

  //       let dv = $("#object-viewer")
  //       dv.empty()

  //       // grab the oembedURI and request embedding info
  //       let oembedUri = event.target.getAttribute('data-uri')
  //       axios.get(oembedUri).then((response)  =>  {
  //         // set a global flag to make the browser think JS jas not all been
  //         // loaded. Without this, the JS file included in the response
  //         // will not load, and the viewer will not render
  //         window.embedScriptIncluded = false
  //         dv.append( $( response.data.html) )
  //         this.$store.commit("setViewerPID", props.model.pid)
  //       }).catch((error) => {
  //         if ( error.message ) {
  //           this.$store.commit("setViewerError", error.message)
  //         } else {
  //           this.$store.commit("setViewerError", error.response.data)
  //         }
  //       }).finally(() => {
  //         this.$store.commit("setViewerLoading", false)
  //       })
  //     }
  //   }
  // }
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
  table.node td.do-buttons {
    text-align: right;
    padding: 5px 2px 10px 0;
  }
  .do-button {
    padding: 5px 25px 4px 25px;
    border-radius: 15px;
    background: #0078e7;
    color: white;
    opacity: 0.7;
    cursor: pointer;
    text-decoration: none;
    margin-left: 5px;
    font-size: 0.9em;
  }
  .do-button:hover {
    opacity: 1;
  }
  .target {
    background: #efffef;
  }
</style>
