<template>
  <div class="content pure-g fixed-header">
    <div class="pure-u-9-24">
      <h4 class="do-header">
        <span>Collection Structure</span>
        <span class="helper-buttons">
          <span class="helper-icon top" @click="scrollTopClick" title="Scroll to top"></span>
          <span class="helper-icon collapse" @click="collapseClick" title="Collapse all"></span>
        </span>
      </h4>
    </div>

    <div class="pure-u-15-24">
      <h4 class="do-header">
        <span>Digitial Object Viewer</span>
        <span v-if='!viewerVisible' class="hint">
          Click 'View Digital Object' from the tree on the left to view it below
        </span>
        <span v-else class="helper-buttons">
          <span class="helper-icon sync" @click="syncClick" title="Sync Tree"></span>
        </span>
      </h4>
    </div>
  </div>
</template>

<script>
  import EventBus from '@/components/EventBus'
  export default {
    name: '',
    data: function () {
      return {
        viewerVisible: false,
        viewerPID: null
      }
    },

    mounted: function () {
      EventBus.$on("viewer-opened", this.handleViewerOpened)
      EventBus.$on('node-destroyed', this.handleNodeDestroyed)
    },

    methods: {
      handleViewerOpened: function(pid) {
        this.viewerVisible = true
        this.viewerPID = pid
      },

      handleNodeDestroyed: function() {
        if (!this.viewerPID) return
        let tgt = $("#"+this.viewerPID)
        if (tgt.length === 0) {
          $("#object-viewer").empty()
          this.viewerPID = null
        }
      },

      scrollTopClick: function() {
        $([document.documentElement, document.body]).animate({
          scrollTop:0
        }, 100);
      },

      collapseClick: function() {
        EventBus.$emit('collapse-all')
        this.scrollTopClick()
      },

      syncClick: function() {
        let tgt = $("#"+this.viewerPID)
        $([document.documentElement, document.body]).animate({
          scrollTop: tgt.offset().top-40
        }, 100);
      },
    }
  }
</script>

<style scoped>
  div.fixed-header {
    background: white;
    z-index: 1000;
    padding-top:5px;
  }
  h4.do-header {
    margin: 0;
    border-bottom: 1px solid #ccc;
    padding-bottom: 10px;
    margin-left: 20px;
    margin-bottom: 0px;
  }
  .hint {
    color: #999;
    margin: 0;
    text-align: right;
    font-size: 0.85em;
    float: right;
    font-weight: 500;
  }
  .helper-icon {
    display: inline-block;
    width:20px;
    height:20px;
    opacity: 0.3;
    cursor: pointer;
    margin-left:5px;
  }
  .helper-icon.collapse {
    background-image: url(../assets/collapse.png);
  }
  .helper-icon.top {
    background-image: url(../assets/top.png);
  }
  .helper-icon.sync {
    background-image: url(../assets/sync.png);
  }
  .helper-buttons {
    float: right;
  }
  .helper-icon:hover {
      opacity: 0.8;
  }
</style>
