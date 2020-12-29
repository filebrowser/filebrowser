import Vue from 'vue'
import { files as api } from '@/api'
import throttle from 'lodash.throttle'
import buttons from '@/utils/buttons'

const UPLOADS_LIMIT = 5;

const state = {
  id: 0,
  size: 0,
  progress: [],
  queue: [],
  uploads: {}
}

const mutations = {
  setProgress(state, { id, loaded }) {
    Vue.set(state.progress, id, loaded)
  },
  reset: (state) => {
    state.id = 0
    state.size = 0
    state.progress = []
  },
  addJob: (state, item) => {
    state.queue.push(item)
    state.size += item.file.size
    state.id++
  },
  moveJob(state) {
    const item = state.queue[0]
    state.queue.shift()
    Vue.set(state.uploads, item.id, item)
  },
  removeJob(state, id) {
    delete state.uploads[id]
  }
}

const beforeUnload = (event) => {
  event.preventDefault()
  event.returnValue = ''
}

const actions = {
  upload: (context, item) => {
    let uploadsCount = Object.keys(context.state.uploads).length;

    let isQueueEmpty = context.state.queue.length == 0
    let isUploadsEmpty = uploadsCount == 0

    if (isQueueEmpty && isUploadsEmpty) {
      window.addEventListener('beforeunload', beforeUnload)
      buttons.loading('upload')
    }

    context.commit('addJob', item)
    context.dispatch('processUploads')
  },
  finishUpload: (context, item) => {
    context.commit('setProgress', { id: item.id, loaded: item.file.size })
    context.commit('removeJob', item.id)
    context.dispatch('processUploads')
  },
  processUploads: async (context) => {
    let uploadsCount = Object.keys(context.state.uploads).length;

    let isBellowLimit = uploadsCount < UPLOADS_LIMIT
    let isQueueEmpty = context.state.queue.length == 0
    let isUploadsEmpty = uploadsCount == 0

    let isFinished = isQueueEmpty && isUploadsEmpty
    let canProcess = isBellowLimit && !isQueueEmpty

    if (isFinished) {
      window.removeEventListener('beforeunload', beforeUnload)
      buttons.success('upload')
      context.commit('reset')
      context.commit('setReload', true, { root: true })
    }

    if (canProcess) {
      const item = context.state.queue[0];
      context.commit('moveJob')

      if (item.file.isDir) {
        await api.post(item.path).catch(Vue.prototype.$showError)
      } else {
        let onUpload = throttle(
          (event) => context.commit('setProgress', { id: item.id, loaded: event.loaded }),
          100, { leading: true, trailing: false }
        )

        await api.post(item.path, item.file, item.overwrite, onUpload).catch(Vue.prototype.$showError)
      }

      context.dispatch('finishUpload', item)
    }
  }
}

export default { state, mutations, actions, namespaced: true }