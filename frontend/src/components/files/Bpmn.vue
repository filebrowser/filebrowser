<template>
  <div
    ref="container"
    class="vue-bpmn-diagram-container"
  />
</template>

<script>
import BpmnJS from 'bpmn-js/dist/bpmn-navigated-viewer.production.min.js';
import { mapState } from 'vuex';
import { baseURL } from '@/utils/constants';

export default {
  name: 'Bpmn',
  data() {
    return {
      url: `${baseURL}/api/raw`,
      diagramXML: null,
      activities: null,
    };
  },
  mounted() {
    const { container } = this.$refs;
    this.bpmnViewer = new BpmnJS({ container });
    const { bpmnViewer, fetchDiagram } = this;
    bpmnViewer.on('import.done', ({ error, warnings }) => {
      if (error) {
        this.$emit('error', error);
      } else {
        this.$emit('shown', warnings);
      }
      bpmnViewer
        .get('canvas')
        .zoom('fit-viewport')
    });
    fetchDiagram();
  },
  beforeDestroy() {
    this.bpmnViewer.destroy();
  },
  watch: {
    url() {
      this.$emit('loading');
      this.fetchDiagram();
    },
    diagramXML(val) {
      this.bpmnViewer.importXML(val);
    }
  },
  methods: {
    fetchDiagram() {
      fetch(`${this.url}/${this.req.path}?auth=${this.jwt}`)
        .then(response => response.text())
        // eslint-disable-next-line no-return-assign
        .then(text => (this.diagramXML = text))
        .catch(err => this.$emit('error', err));
    }
  },
  computed: mapState(['data', 'req', 'jwt']),
};
</script>

<style>
  .vue-bpmn-diagram-container {
    height: 83vh;
    width: 100%;
  }
</style>
