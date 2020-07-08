import store from '@/store'
import { files as api } from '@/api'
import throttle from 'lodash.throttle'
import buttons from '@/utils/buttons'
import url from '@/utils/url'

export function checkConflict(files, items) {
  if (typeof items === 'undefined' || items === null) {
    items = []
  }

  let folder_upload = false
  if (files[0].fullPath !== undefined) {
    folder_upload = true
  }

  let conflict = false
  for (let i = 0; i < files.length; i++) {
    let file = files[i]
    let name = file.name

    if (folder_upload) {
      let dirs = file.fullPath.split("/")
      if (dirs.length > 1) {
        name = dirs[0]
      }
    }

    let res = items.findIndex(function hasConflict(element) {
      return (element.name === this)
    }, name)

    if (res >= 0) {
      conflict = true
      break
    }
  }

  return conflict
}

export function scanFiles(dt) {
  return new Promise((resolve) => {
    let reading = 0
    const contents = []

    if (dt.items !== undefined) {
      for (let item of dt.items) {
        if (item.kind === "file" && typeof item.webkitGetAsEntry === "function") {
          const entry = item.webkitGetAsEntry()
          readEntry(entry)
        }
      }
    } else {
      resolve(dt.files)
    }

    function readEntry(entry, directory = "") {
      if (entry.isFile) {
        reading++
        entry.file(file => {
          reading--

          file.fullPath = `${directory}${file.name}`
          contents.push(file)

          if (reading === 0) {
            resolve(contents)
          }
        })
      } else if (entry.isDirectory) {
        const dir = {
          isDir: true,
          path: `${directory}${entry.name}`
        }

        contents.push(dir)

        readReaderContent(entry.createReader(), `${directory}${entry.name}`)
      }
    }

    function readReaderContent(reader, directory) {
      reading++

      reader.readEntries(function (entries) {
        reading--
        if (entries.length > 0) {
          for (const entry of entries) {
            readEntry(entry, `${directory}/`)
          }

          readReaderContent(reader, `${directory}/`)
        }

        if (reading === 0) {
          resolve(contents)
        }
      })
    }
  })
}

export function handleFiles(files, path, overwrite = false) {
  if (store.state.upload.count == 0) {
    buttons.loading('upload')
  }

  let promises = []

  let onupload = (id) => (event) => {
    store.commit('upload/setProgress', { id, loaded: event.loaded })
  }

  for (let i = 0; i < files.length; i++) {
    let file = files[i]

    if (!file.isDir) {
      let filename = (file.fullPath !== undefined) ? file.fullPath : file.name
      let filenameEncoded = url.encodeRFC5987ValueChars(filename)

      let id = store.state.upload.id

      store.commit('upload/incrementSize', file.size)
      store.commit('upload/incrementId')
      store.commit('upload/incrementCount')

      let promise = api.post(path + filenameEncoded, file, overwrite, throttle(onupload(id), 100)).finally(() => {
        store.commit('upload/decreaseCount')
      })

      promises.push(promise)
    } else {
      let uri = path
      let folders = file.path.split("/")

      for (let i = 0; i < folders.length; i++) {
        let folder = folders[i]
        let folderEncoded = encodeURIComponent(folder)
        uri += folderEncoded + "/"
      }

      api.post(uri)
    }
  }

  let finish = () => {
    if (store.state.upload.count > 0) {
      return
    }

    buttons.success('upload')

    store.commit('setReload', true)
    store.commit('upload/reset')
  }

  Promise.all(promises)
    .then(() => {
      finish()
    })
    .catch(error => {
      finish()
      this.$showError(error)
    })

  return false
}