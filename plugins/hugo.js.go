package plugins

const hugoJavaScript = `'use strict';

(function () {
  if (window.plugins === undefined || window.plugins === null) {
    window.plugins = []
  }

  let regenerate = function (data, url) {
    url = data.api.removePrefix(url)

    return new Promise((resolve, reject) => {
      let request = new window.XMLHttpRequest()
      request.open('POST', data.store.state.baseURL + "/api/hugo" + url, true)
      request.setRequestHeader('Authorization', "Bearer " + data.store.state.jwt)
      request.setRequestHeader('Regenerate', 'true')

      request.onload = () => {
        if (request.status === 200) {
          resolve()
        } else {
          reject(request.responseText)
        }
      }

      request.onerror = (error) => reject(error)
      request.send()
    })
  }

  let newArchetype = function (data, url, type) {
    url = data.api.removePrefix(url)

    return new Promise((resolve, reject) => {
      let request = new window.XMLHttpRequest()
      request.open('POST', data.store.state.baseURL + "/api/hugo" + url, true)
      request.setRequestHeader('Authorization',"Bearer " + data.store.state.jwt)
      request.setRequestHeader('Archetype', encodeURIComponent(type))

      request.onload = () => {
        if (request.status === 200) {
          resolve(request.getResponseHeader('Location'))
        } else {
          reject(request.responseText)
        }
      }

      request.onerror = (error) => reject(error)
      request.send()
    })
  }

  let schedule = function (data, file, date) {
    file = data.api.removePrefix(file)

    return new Promise((resolve, reject) => {
      let request = new window.XMLHttpRequest()
      request.open('POST', data.store.state.baseURL + "/api/hugo" + file, true)
      request.setRequestHeader('Authorization', "Bearer " + data.store.state.jwt)
      request.setRequestHeader('Schedule', date)

      request.onload = () => {
        if (request.status === 200) {
          resolve(request.getResponseHeader('Location'))
        } else {
          reject(request.responseText)
        }
      }

      request.onerror = (error) => reject(error)
      request.send()
    })
  }

  window.plugins.push({
    name: 'hugo',
    credits: 'With a flavour of <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-hugo">Hugo</a>.',
    header: {
      visible: [
        {
          if: function (data, route) {
            return (data.store.state.req.kind === 'editor' &&
              !data.store.state.loading &&
              data.store.state.user.allowEdit &
              data.store.state.user.permissions.allowPublish)
          },
          click: function (event, data, route) {
            event.preventDefault()
            document.getElementById('save-button').click()
            // TODO: wait for save to finish?
            data.buttons.loading('publish')

            regenerate(data, route.path)
              .then(() => {
                data.buttons.done('publish')
                data.store.commit('showSuccess', 'Post published!')
                data.store.commit('setReload', true)
              })
              .catch((error) => {
                data.buttons.done('publish')
                data.store.commit('showError', error)
              })
          },
          id: 'publish-button',
          icon: 'send',
          name: 'Publish'
        }
      ],
      hidden: [
        {
          if: function (data, route) {
            return (data.store.state.req.kind === 'editor' &&
              !data.store.state.loading &&
              data.store.state.req.metadata !== undefined &&
              data.store.state.req.metadata !== null &&
              data.store.state.user.permissions.allowPublish)
          },
          click: function (event, data, route) {
            document.getElementById('save-button').click()
            data.store.commit('showHover', 'schedule')
          },
          id: 'schedule-button',
          icon: 'alarm',
          name: 'Schedule'
        }
      ]
    },
    sidebar: [
      {
        click: function (event, data, route) {
          data.router.push({ path: '/files/settings' })
        },
        icon: 'settings',
        name: 'Hugo Settings'
      },
      {
        click: function (event, data, route) {
          data.store.commit('showHover', 'new-archetype')
        },
        if: function (data, route) {
          return data.store.state.user.allowNew
        },
        icon: 'merge_type',
        name: 'Hugo New'
      },
      {
        click: function (event, data, route) {
          window.open(data.store.state.baseURL + '/preview/')
        },
        icon: 'remove_red_eye',
        name: 'Preview'
      }
    ],
    prompts: [
      {
        name: 'new-archetype',
        title: 'New file',
        description: 'Create a new post based on an archetype. Your file will be created on content folder.',
        inputs: [
          {
            type: 'text',
            name: 'file',
            placeholder: 'File name'
          },
          {
            type: 'text',
            name: 'archetype',
            placeholder: 'Archetype'
          }
        ],
        ok: 'Create',
        submit: function (event, data, route) {
          event.preventDefault()

          let file = event.currentTarget.querySelector('[name="file"]').value
          let type = event.currentTarget.querySelector('[name="archetype"]').value
          if (type === '') type = 'default'

          data.store.commit('closeHovers')

          newArchetype(data, '/' + file, type)
            .then((url) => {
              data.router.push({ path: url })
            })
            .catch(error => {
              data.store.commit('showError', error)
            })
        }
      },
      {
        name: 'schedule',
        title: 'Schedule',
        description: 'Pick a date and time to schedule the publication of this post.',
        inputs: [
          {
            type: 'datetime-local',
            name: 'date',
            placeholder: 'Date'
          }
        ],
        ok: 'Schedule',
        submit: function (event, data, route) {
          event.preventDefault()
          data.buttons.loading('schedule')

          let date = event.currentTarget.querySelector('[name="date"]').value
          if (date === '') {
            data.buttons.done('schedule')
            data.store.commit('showError', 'The date must not be empty.')
            return
          }

          schedule(data, route.path, date)
            .then(() => {
              data.buttons.done('schedule')
              data.store.commit('showSuccess', 'Post scheduled!')
            })
            .catch((error) => {
              data.buttons.done('schedule')
              data.store.commit('showError', error)
            })
        }
      }
    ]
  })
})()`
