<template>
   <li class="tree-node">
      <div class="controls">
         <span v-if="isFolder" class="icon" @click="toggle" :class="{ plus: isOpen == false, minus: isOpen == true }"></span>
         <span v-if="isEditing" class="editing-ctls">
            <span class="do-button" @click="cancelEdit()">Cancel</span>
            <span class="do-button" @click="submitEdit()">Submit</span>
         </span>
         <span v-else class="edit do-button" @click="editNode()"
            :class="{ disabled: collectionStore.editParentPID !='' }">Edit</span>
      </div>
      <table class="node" :id="model.pid">
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
            <template v-if="attribute.type.name != 'digitalObject'">
               <td class="label">{{ attribute.type.name }}:</td>
               <td class="data">
                  <input v-if="isEditing && attribute.type.name == 'title'" type="text" v-model="newTitle">
                  <textarea v-else-if="isEditing && attribute.type.name == 'description'" type="text" v-model="newDesc" :rows="8"></textarea>
                  <span v-else v-html="renderAttributeValue(attribute)"></span>
               </td>
            </template>
            <template v-else>
               <td colspan="2" class="do-buttons">
                  <a v-if="hasIIIFManifest" class="do-button" :href="iiifManufestURL" target="_blank">IIIF Manifest</a>
                  <span @click="digitalObjectClicked(model.pid, attribute.values[0].value)" class="do-button">View Digitial
                     Object</span>
               </td>
            </template>
         </tr>
      </table>
      <ul v-if="isOpen">
         <CollectionDetailsItem v-for="child in model.children" :key="child.pid" :model="child" :open="false" />
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
   open: {
      type: Boolean,
      default: false
   },
})

const newTitle = ref("")
const newDesc = ref("")

const isEditing = computed(() => {
   return collectionStore.editParentPID == props.model.pid
})
const isOpen = computed(() => {
   return props.model.open
})
const isFolder = computed(() => {
   return props.model.children && props.model.children.length
})
const iiifManufestURL = computed(() => {
   return IIIF_MAN_URL + "/" + externalPID() + "/manifest.json"
})
const hasIIIFManifest = computed(() => {
   if (props.model.type.name === "item") {
      return false
   }
   return true
})

const editNode = (() => {
   if (collectionStore.editParentPID != "") return
   collectionStore.startEdit(props.model.pid)
   props.model.attributes.forEach(a => {
      if (a.type.name == 'title') {
         newTitle.value = a.values[0].value
      }
      if (a.type.name == 'description') {
         newDesc.value = a.values[0].value
      }
   })
})
const cancelEdit = (() => {
   collectionStore.cancelEdit()
})
const submitEdit = (() => {
   collectionStore.submitEdit(newTitle.value, newDesc.value)
})

const toggle = (() => {
   collectionStore.toggleOpen(props.model.pid)
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

const renderAttributeValue = ((attribute) => {
   let out = ""
   for (var idx in attribute.values) {
      let val = attribute.values[idx]
      if (out.length > 0) out += "<span>, </span>"
      if (val.valueURI) {
         out += "<a class='uri' href='" + val.valueURI + "' target='_blank'>" + val.value + "</a>"
      } else {
         out += "<span>" + val.value + "</span>"
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

   td {
      padding: 5px 10px 5px 0;
      input,textarea {
         width: 100%;
         color: #444;
      }
   }
   tr {
      padding: 5px 10px 5px 0;
   }
   td.label {
      text-transform: capitalize;
      margin-right: 15px;
      font-weight: bold;
      text-align: right;
      padding: 5px 10px 5px 10px;
      width: 35%;
      min-width: 90px;
      vertical-align: baseline;
   }
}

table.node.selected {
   border-left: 2px solid #7af;
   border-bottom: 2px solid #7af;
   background: #f0f7ff;
}

div.content ul,
div.content li {
   font-size: 14px;
   list-style-type: none;
   position: relative;
}

.controls {
   display: flex;
   flex-flow: row nowrap;
   justify-content: space-between;
   padding: 5px;
   border-left: 1px solid #ccc;
   border-bottom: 1px solid #ccc;
   align-items: center;
   background: #efefff;

   .edit, .editing-ctls {
      margin-left: auto;
   }

   span.icon {
      width: 18px;
      height: 18px;
      padding: 0;
      z-index: 100;
      cursor: pointer;
      opacity: 0.4;
      margin-left: 3px;
      background-repeat: no-repeat;
      background-position: 0;

      &:hover {
         opacity: 0.7;
      }
   }

   span.icon.plus {
      background: url(../assets/plus.png);
   }

   span.icon.minus {
      background: url(../assets/minus.png);
   }

   .do-button {
      font-size: 0.85em;
      border-radius: 3px;
      background: #888;
      padding: 3px 8px;
      font-weight: bold;
      opacity: 0.8;
      cursor: pointer;
      display: inline-block;
      z-index: 9999;

      &:hover {
         opacity: 1.0;
      }
   }

   .do-button.disabled {
      opacity: 0.3;
      cursor: default;

      &:hover {
         opacity: 0.3;
      }
   }
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
   &:hover {
      opacity: 1;
   }
}
</style>
