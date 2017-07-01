'use strict'

if (window.plugins === undefined || window.plugins === null) {
  window.plugins = []
}

window.plugins.append({
  sidebar: [
    {
      click: function (event) {
        console.log('evt')
      },
      icon: 'settings_applications',
      name: 'Settings'
    },
    {
      click: function (event) {
        console.log('evt')
      },
      icon: 'remove_red_eye',
      name: 'Preview'
    }
  ]
})


/*
{{ define "sidebar-addon" }}
<a class="action" href="{{ .BaseURL }}/content/">
    <i class="material-icons">subject</i>
    <span>Posts and Pages</span>
</a>
<a class="action" href="{{ .BaseURL }}/themes/">
    <i class="material-icons">format_paint</i>
    <span>Themes</span>
</a>
<a class="action" href="{{ .BaseURL }}/settings/">
    <i class="material-icons">settings</i>
    <span>Settings</span>
</a>
{{ end }}
*/
