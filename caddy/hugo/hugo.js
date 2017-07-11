'use strict';

(function () {
  if (window.plugins === undefined || window.plugins === null) {
    window.plugins = []
  }

  let regenerate = function (data, url) {
    url = data.api.removePrefix(url)

    return new Promise((resolve, reject) => {
      let request = new window.XMLHttpRequest()
      request.open('POST', `${data.store.state.baseURL}/api/hugo${url}`, true)
      request.setRequestHeader('Authorization', `Bearer ${data.store.state.jwt}`)
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

  let newArchetype = function (data, file, type) {
    file = data.api.removePrefix(file)

    return new Promise((resolve, reject) => {
      let request = new window.XMLHttpRequest()
      request.open('POST', `${data.store.state.baseURL}/api/hugo${file}`, true)
      request.setRequestHeader('Authorization', `Bearer ${data.store.state.jwt}`)
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

  window.plugins.push({
    name: 'hugo',
    credits: 'With a flavour of <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-hugo">Hugo</a>.',
    header: {
      visible: [
        {
          if: function (data, route) {
            return (data.store.state.req.kind === 'editor' &&
              !data.store.state.loading &&
              data.store.state.req.metadata !== undefined &&
              data.store.state.req.metadata !== null &&
              data.store.state.user.allowEdit)
            // TODO: add allowPublish
          },
          click: function (event, data, route) {
            event.preventDefault()
            document.getElementById('save-button').click()
            // TODO: wait for save to finish?
            data.buttons.loading('publish')

            regenerate(data, route.path)
              .then(() => {
                data.buttons.done('publish')
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
              data.store.state.req.metadata !== null)
          },
          click: function (event, data, route) {
            console.log('Schedule')
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
        icon: 'settings_applications',
        name: 'Settings'
      },
      {
        click: function (event, data, route) {
          data.store.commit('showHover', 'new-archetype')
        },
        icon: 'merge_type',
        name: 'Hugo new'
      },
      {
        click: function (event, data, route) {
          console.log('evt')
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

          console.log(event)

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
      }
    ]
  })
})()
