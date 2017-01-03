'use strict';

document.addEventListener('DOMContentLoaded', event => {
  document.querySelector('#top-bar > div > p:first-child').innerHTML = 'Hugo for Caddy'
  document.querySelector('footer').innerHTML += ' With a flavour of <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-hugo">Hugo</a>.';

  document.querySelector('#bottom-bar>*:first-child').style.maxWidth = "calc(100% - 27em)"

  let link = baseURL + "/settings/"

  document.getElementById('info').insertAdjacentHTML('beforebegin', `<a href="${link}">
      <div class="action">
       <i class="material-icons">settings</i>
       <span>Settings</span>
      </div>
     </a>`);

  if(buttons.new && window.location.pathname === baseURL + "/content/") {
    buttons.new.removeEventListener('click', listing.newFileButton);
    buttons.new.addEventListener('click', hugo.newFileButton);
  }

  if(buttons.save) {
    let box = document.getElementById('file-only');

    box.insertAdjacentHTML('beforeend', `<div class="action" id="publish">
           <i class="material-icons">send</i>
           <span>Publish</span>
          </div>`);

    buttons.publish = document.getElementById('publish')
    buttons.publish.addEventListener('click', hugo.publish)

    if((document.getElementById('date') || document.getElementById('publishdate')) &&
      document.getElementById('editor').dataset.kind == "complete") {

      box.insertAdjacentHTML('beforeend', ` <div class="action" id="schedule">
             <i class="material-icons">alarm</i>
             <span>Schedule</span>
            </div>`);

      buttons.schedule = document.getElementById('schedule')
      buttons.schedule.addEventListener('click', hugo.schedule)
    }

    document.querySelector('#bottom-bar>*:first-child').style.maxWidth = "calc(100% - 30em)"
  }
});

var hugo = {};

hugo.newFileButton = function (event) {
  event.preventDefault();
  event.stopPropagation();

  let clone = document.importNode(templates.question.content, true);
  clone.querySelector('h3').innerHTML = 'New file';
  clone.querySelector('p').innerHTML = 'End with a trailing slash to create a dir. To use an archetype, use <code>file[:archetype]</code>.';
  clone.querySelector('.ok').innerHTML = 'Create';
  clone.querySelector('form').addEventListener('submit', hugo.newFilePrompt);

  document.querySelector('body').appendChild(clone)
  document.querySelector('.overlay').classList.add('active');
  document.querySelector('.prompt').classList.add('active');
}

hugo.newFilePrompt = function (event) {
  event.preventDefault();
  buttons.setLoading('new');

  let value = event.currentTarget.querySelector('input').value,
    index = value.lastIndexOf(':'),
    name = value.substring(0, index),
    archetype = value.substring(index + 1, value.length);

  if(name == "") name = archetype;
  if(index == -1) archetype = "";

  webdav.new(window.location.pathname + name, '', {
      'Filename': name,
      'Archetype': archetype
    })
    .then(() => {
      buttons.setDone('new');
      window.location = window.location.pathname + name;
    })
    .catch(e => {
      console.log(e);
      buttons.setDone('new', false);
    });

  closePrompt(event);
  return false;
}

hugo.publish = function (event) {
  event.preventDefault();

  if(document.getElementById('draft')) {
    document.getElementById('block-draft').remove();
  }

  buttons.setLoading('publish');

  let data = JSON.stringify(form2js(document.querySelector('form'))),
    headers = {
      'Kind': document.getElementById('editor').dataset.kind,
      'Regenerate': 'true'
    };

  webdav.put(window.location.pathname, data, headers)
    .then(() => {
      buttons.setDone('publish');
    })
    .catch(e => {
      console.log(e);
      buttons.setDone('publish', false)
    })
}

hugo.schedule = function (event) {
  event.preventDefault();

  let date = document.getElementById('date').value;
  if(document.getElementById('publishDate')) {
    date = document.getElementById('publishDate').value;
  }

  buttons.setLoading('publish');

  let data = JSON.stringify(form2js(document.querySelector('form'))),
    headers = {
      'Kind': document.getElementById('editor').dataset.kind,
      'Schedule': 'true'
    };

  webdav.put(window.location.pathname, data, headers)
    .then(() => {
      buttons.setDone('publish');
    })
    .catch(e => {
      console.log(e);
      buttons.setDone('publish', false)
    })
}
