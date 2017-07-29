package plugins

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    `hugo.js`,
		FileModTime: time.Unix(1501319450, 0),
		Content:     string("'use strict';\n\n(function () {\n  if (window.plugins === undefined || window.plugins === null) {\n    window.plugins = []\n  }\n\n  let regenerate = function (data, url) {\n    url = data.api.removePrefix(url)\n\n    return new Promise((resolve, reject) => {\n      let request = new window.XMLHttpRequest()\n      request.open('POST', `${data.store.state.baseURL}/api/hugo${url}`, true)\n      request.setRequestHeader('Authorization', `Bearer ${data.store.state.jwt}`)\n      request.setRequestHeader('Regenerate', 'true')\n\n      request.onload = () => {\n        if (request.status === 200) {\n          resolve()\n        } else {\n          reject(request.responseText)\n        }\n      }\n\n      request.onerror = (error) => reject(error)\n      request.send()\n    })\n  }\n\n  let newArchetype = function (data, url, type) {\n    url = data.api.removePrefix(url)\n\n    return new Promise((resolve, reject) => {\n      let request = new window.XMLHttpRequest()\n      request.open('POST', `${data.store.state.baseURL}/api/hugo${url}`, true)\n      request.setRequestHeader('Authorization', `Bearer ${data.store.state.jwt}`)\n      request.setRequestHeader('Archetype', encodeURIComponent(type))\n\n      request.onload = () => {\n        if (request.status === 200) {\n          resolve(request.getResponseHeader('Location'))\n        } else {\n          reject(request.responseText)\n        }\n      }\n\n      request.onerror = (error) => reject(error)\n      request.send()\n    })\n  }\n\n  let schedule = function (data, file, date) {\n    file = data.api.removePrefix(file)\n\n    return new Promise((resolve, reject) => {\n      let request = new window.XMLHttpRequest()\n      request.open('POST', `${data.store.state.baseURL}/api/hugo${file}`, true)\n      request.setRequestHeader('Authorization', `Bearer ${data.store.state.jwt}`)\n      request.setRequestHeader('Schedule', date)\n\n      request.onload = () => {\n        if (request.status === 200) {\n          resolve(request.getResponseHeader('Location'))\n        } else {\n          reject(request.responseText)\n        }\n      }\n\n      request.onerror = (error) => reject(error)\n      request.send()\n    })\n  }\n\n  window.plugins.push({\n    name: 'hugo',\n    credits: 'With a flavour of <a rel=\"noopener noreferrer\" href=\"https://github.com/hacdias/caddy-hugo\">Hugo</a>.',\n    header: {\n      visible: [\n        {\n          if: function (data, route) {\n            return (data.store.state.req.kind === 'editor' &&\n              !data.store.state.loading &&\n              data.store.state.user.allowEdit &\n              data.store.state.user.permissions.allowPublish)\n          },\n          click: function (event, data, route) {\n            event.preventDefault()\n            document.getElementById('save-button').click()\n            // TODO: wait for save to finish?\n            data.buttons.loading('publish')\n\n            regenerate(data, route.path)\n              .then(() => {\n                data.buttons.done('publish')\n                data.store.commit('showSuccess', 'Post published!')\n                data.store.commit('setReload', true)\n              })\n              .catch((error) => {\n                data.buttons.done('publish')\n                data.store.commit('showError', error)\n              })\n          },\n          id: 'publish-button',\n          icon: 'send',\n          name: 'Publish'\n        }\n      ],\n      hidden: [\n        {\n          if: function (data, route) {\n            return (data.store.state.req.kind === 'editor' &&\n              !data.store.state.loading &&\n              data.store.state.req.metadata !== undefined &&\n              data.store.state.req.metadata !== null &&\n              data.store.state.user.permissions.allowPublish)\n          },\n          click: function (event, data, route) {\n            document.getElementById('save-button').click()\n            data.store.commit('showHover', 'schedule')\n          },\n          id: 'schedule-button',\n          icon: 'alarm',\n          name: 'Schedule'\n        }\n      ]\n    },\n    sidebar: [\n      {\n        click: function (event, data, route) {\n          data.router.push({ path: '/files/settings' })\n        },\n        icon: 'settings',\n        name: 'Hugo Settings'\n      },\n      {\n        click: function (event, data, route) {\n          data.store.commit('showHover', 'new-archetype')\n        },\n        if: function (data, route) {\n          return data.store.state.user.allowNew\n        },\n        icon: 'merge_type',\n        name: 'Hugo New'\n      } /* ,\n      {\n        click: function (event, data, route) {\n          console.log('evt')\n        },\n        icon: 'remove_red_eye',\n        name: 'Preview'\n      } */\n    ],\n    prompts: [\n      {\n        name: 'new-archetype',\n        title: 'New file',\n        description: 'Create a new post based on an archetype. Your file will be created on content folder.',\n        inputs: [\n          {\n            type: 'text',\n            name: 'file',\n            placeholder: 'File name'\n          },\n          {\n            type: 'text',\n            name: 'archetype',\n            placeholder: 'Archetype'\n          }\n        ],\n        ok: 'Create',\n        submit: function (event, data, route) {\n          event.preventDefault()\n\n          let file = event.currentTarget.querySelector('[name=\"file\"]').value\n          let type = event.currentTarget.querySelector('[name=\"archetype\"]').value\n          if (type === '') type = 'default'\n\n          data.store.commit('closeHovers')\n\n          newArchetype(data, '/' + file, type)\n            .then((url) => {\n              data.router.push({ path: url })\n            })\n            .catch(error => {\n              data.store.commit('showError', error)\n            })\n        }\n      },\n      {\n        name: 'schedule',\n        title: 'Schedule',\n        description: 'Pick a date and time to schedule the publication of this post.',\n        inputs: [\n          {\n            type: 'datetime-local',\n            name: 'date',\n            placeholder: 'Date'\n          }\n        ],\n        ok: 'Schedule',\n        submit: function (event, data, route) {\n          event.preventDefault()\n          data.buttons.loading('schedule')\n\n          let date = event.currentTarget.querySelector('[name=\"date\"]').value\n          if (date === '') {\n            data.buttons.done('schedule')\n            data.store.commit('showError', 'The date must not be empty.')\n            return\n          }\n\n          schedule(data, route.path, date)\n            .then(() => {\n              data.buttons.done('schedule')\n              data.store.commit('showSuccess', 'Post scheduled!')\n            })\n            .catch((error) => {\n              data.buttons.done('schedule')\n              data.store.commit('showError', error)\n            })\n        }\n      }\n    ]\n  })\n})()\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   ``,
		DirModTime: time.Unix(1501318911, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // hugo.js

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`./assets/`, &embedded.EmbeddedBox{
		Name: `./assets/`,
		Time: time.Unix(1501318911, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"hugo.js": file2,
		},
	})
}
