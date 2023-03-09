<template>
  <li class="tree-node">
    <span v-if="isFolder" class="icon" @click="toggle" :class="{ plus: isOpen==false, minus: isOpen==true}"></span>
    <table class="node" :id="model.pid" >
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
            <span @click="digitalObjectClicked(model.pid, attribute.values[0].value)" class="do-button">View Digitial Object</span>
          </td>
        </template>
      </tr>
    </table>
    <ul v-if="isOpen">
      <CollectionDetailsItem  v-for="child in model.children" :key="child.pid" :model="child" :depth="depth+1" :open="false"/>
    </ul>
  </li>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useCollectionsStore } from '@/stores/collections'

const IIIF_MAN_URL = import.meta.env.VITE_IIIF_MAN_URL

const collectionStore = useCollectionsStore()

const props = defineProps({
   model: {
      type: Object,
      required: true
   },
   depth: {
      type: Number,
      required: true
   },
   open: {
      type: Boolean,
      default: false
   },
})
const isOpen = computed( () => {
   return props.model.open
})
const isFolder = computed( () => {
   return props.model.children && props.model.children.length
})
const iiifManufestURL = computed( () => {
   return IIIF_MAN_URL+"/"+externalPID()+"/manifest.json"
})
const hasIIIFManifest = computed( () => {
   if (props.model.type.name === "item") {
      return false
   }
   return true
})

const toggle = (() => {
  collectionStore.toggleOpen( props.model.pid )
})

const digitalObjectClicked = ((pid, viewerAttribString) => {
   // make sure only one node is marked as selected
   let nodes = document.getElementsByClassName("node")
   for (let n of nodes) {
      n.classList.remove("selected")
   }
   let tgtNode = document.getElementById(pid)
   tgtNode.classList.add("selected")

   let attrib = JSON.parse(viewerAttribString)
   let externalID = attrib.id

   let viewewrDiv = document.getElementById("object-viewer")
   collectionStore.loadViewer(viewewrDiv, pid, externalID)
})

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
</script>

<style scoped lang="scss">
.tree-node {
   list-style: none;
   position: relative;
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
