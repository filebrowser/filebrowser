import Vue from 'vue'

const state = {
  id: 0,
  count: 0,
  size: 0,
  progress: []
}

const mutations = {
  incrementId: (state) => {
    state.id = state.id + 1
  },
  incrementSize: (state, value) => {
    state.size = state.size + value
  },
  incrementCount: (state) => {
    state.count = state.count + 1
  },
  decreaseCount: (state) => {
    state.count = state.count - 1
  },
  setProgress(state, { id, loaded }) {
    Vue.set(state.progress, id, loaded)
  },
  reset: (state) => {
    state.id = 0
    state.size = 0
    state.count = 0
    state.progress = []
  }
}

export default { state, mutations, namespaced: true }