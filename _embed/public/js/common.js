'use strict';

var tempID = "_fm_internal_temporary_id",
  ssl = (location.protocol === 'https:'),
  templates = {},
  selectedItems = [],
  overlay,
  clickOverlay;

// Removes an element, if exists, from an array
Array.prototype.removeElement = function (element) {
  var i = this.indexOf(element);
  if(i != -1)
    this.splice(i, 1);
}

// Replaces an element inside an array by another
Array.prototype.replaceElement = function (oldElement, newElement) {
  var i = this.indexOf(oldElement);
  if(i != -1)
    this[i] = newElement;
}

// Sends a costum event to itself
Document.prototype.sendCostumEvent = function (text) {
  this.dispatchEvent(new CustomEvent(text));
}

// Gets the content of a cookie
Document.prototype.getCookie = function (name) {
  var re = new RegExp("(?:(?:^|.*;\\s*)" + name + "\\s*\\=\\s*([^;]*).*$)|^.*$");
  return document.cookie.replace(re, "$1");
}

// Remove the last directory of an url
var removeLastDirectoryPartOf = function (url) {
  var arr = url.split('/');
  if(arr.pop() === "") {
    arr.pop();
  }
  return(arr.join('/'));
}

function getCSSRule(rules) {
  for(let i = 0; i < rules.length; i++) {
    rules[i] = rules[i].toLowerCase();
  }

  var result = null,
    find = Array.prototype.find;

  find.call(document.styleSheets, styleSheet => {
    result = find.call(styleSheet.cssRules, cssRule => {
      let found = false;

      if(cssRule instanceof CSSStyleRule) {
        for(let i = 0; i < rules.length; i++) {
          if(cssRule.selectorText.toLowerCase() == rules[i])
            found = true;
        }
      }

      return found;
    });

    return result != null;
  });

  return result;
}

/* * * * * * * * * * * * * * * *
 *                             *
 *            BUTTONS          *
 *                             *
 * * * * * * * * * * * * * * * */
var buttons = {
  previousState: {},
};

buttons.setLoading = function (name) {
  if(typeof this[name] === 'undefined') return;
  let i = this[name].querySelector('i');

  this.previousState[name] = i.innerHTML;
  i.style.opacity = 0;

  setTimeout(function () {
    i.classList.add('spin');
    i.innerHTML = 'autorenew';
    i.style.opacity = 1;
  }, 200);
}

// Changes an element to done animation
buttons.setDone = function (name, success = true) {
  let i = this[name].querySelector('i');

  i.style.opacity = 0;

  let thirdStep = () => {
    i.innerHTML = this.previousState[name];
    i.style.opacity = null;

    if(selectedItems.length == 0 && document.getElementById('listing'))
      document.sendCostumEvent('changed-selected');
  }

  let secondStep = () => {
    i.style.opacity = 0;
    setTimeout(thirdStep, 200);
  }

  let firstStep = () => {
    i.classList.remove('spin');
    i.innerHTML = success ?
      'done' :
      'close';
    i.style.opacity = 1;
    setTimeout(secondStep, 1000);
  }

  setTimeout(firstStep, 200);
  return false;
}

/* * * * * * * * * * * * * * * *
 *                             *
 *            WEBDAV           *
 *                             *
 * * * * * * * * * * * * * * * */
var webdav = {};

webdav.convertURL = function (url) {
  return window.location.origin + url.replace(baseURL + "/", webdavURL + "/");
}

webdav.move = function (oldLink, newLink) {
  return new Promise((resolve, reject) => {
    let request = new XMLHttpRequest();
    request.open('MOVE', webdav.convertURL(oldLink), true);
    request.setRequestHeader('Destination', webdav.convertURL(newLink));
    request.onload = () => {
      if(request.status == 201 || request.status == 204) {
        resolve();
      } else {
        reject(request.statusText);
      }
    }
    request.onerror = () => reject(request.statusText);
    request.send();
  });
}

webdav.put = function (link, body, headers = {}) {
  return new Promise((resolve, reject) => {
    let request = new XMLHttpRequest();
    request.open('PUT', webdav.convertURL(link), true);

    for(let key in headers) {
      request.setRequestHeader(key, headers[key]);
    }

    request.onload = () => {
      if(request.status == 201) {
        resolve();
      } else {
        reject(request.statusText);
      }
    }
    request.onerror = () => reject(request.statusText);
    request.send(body);
  });
}

webdav.delete = function (link) {
  return new Promise((resolve, reject) => {
    let request = new XMLHttpRequest();
    request.open('DELETE', webdav.convertURL(link), true);
    request.onload = () => {
      if(request.status == 204) {
        resolve();
      } else {
        reject(request.statusText);
      }
    }
    request.onerror = () => reject(request.statusText);
    request.send();
  });
  request.send();
}

webdav.new = function (link) {
  return new Promise((resolve, reject) => {
    let request = new XMLHttpRequest();
    request.open((link.endsWith("/") ? "MKCOL" : "PUT"), webdav.convertURL(link), true);
    request.onload = () => {
      if(request.status == 201) {
        resolve();
      } else {
        reject(request.statusText);
      }
    }
    request.onerror = () => reject(request.statusText);
    request.send();
  });
  request.send();
}

/* * * * * * * * * * * * * * * *
 *                             *
 *            EVENTS           *
 *                             *
 * * * * * * * * * * * * * * * */
function closePrompt(event) {
  let prompt = document.querySelector('.prompt');

  if(!prompt)
    return;

  event.preventDefault();
  document.querySelector('.overlay').classList.remove('active');
  prompt.classList.remove('active');

  setTimeout(() => {
    prompt.remove();
  }, 100);
}

function notImplemented(event) {
  event.preventDefault();
  clickOverlay.click();

  let clone = document.importNode(templates.info.content, true);
  clone.querySelector('h3').innerHTML = 'Not implemented';
  clone.querySelector('p').innerHTML = "Sorry, but this feature wasn't implemented yet.";

  document.querySelector('body').appendChild(clone)
  document.querySelector('.overlay').classList.add('active');
  document.querySelector('.prompt').classList.add('active');
}

// Prevent Default event
var preventDefault = function (event) {
  event.preventDefault();
}

function logoutEvent(event) {
  let request = new XMLHttpRequest();
  request.open('GET', window.location.pathname, true, "username", "password");
  request.send();
  request.onreadystatechange = function () {
    if(request.readyState == 4) {
      window.location = "/";
    }
  }
}

function openEvent(event) {
  if(event.currentTarget.classList.contains('disabled'))
    return false;

  let link = '?raw=true';

  if(selectedItems.length) {
    link = document.getElementById(selectedItems[0]).dataset.url + link;
  } else {
    link = window.location + link;
  }

  window.open(link);
  return false;
}

function selectMoveFolder(event) {
  if(event.target.getAttribute("aria-selected") === "true") {
    event.target.setAttribute("aria-selected", false);
    return;
  } else {
    if(document.querySelector(".file-list li[aria-selected=true]")) {
      document.querySelector(".file-list li[aria-selected=true]").setAttribute("aria-selected", false);
    }
    event.target.setAttribute("aria-selected", true);
    return;
  }
}

function loadNextFolder(event) {
  let request = new XMLHttpRequest(),
    prompt = document.querySelector("form.prompt.active");

  prompt.addEventListener("submit", moveSelected);

  request.open("GET", event.target.dataset.url);
  request.setRequestHeader("Accept", "application/json");
  request.send();
  request.onreadystatechange = function () {
    if(request.readyState == 4 && request.status == 200) {
      let dirs = 0;

      prompt.querySelector("ul").innerHTML = "";
      prompt.querySelector('code').innerHTML = event.target.dataset.url;

      if(JSON.parse(request.response) == null) {
        prompt.querySelector("p").innerHTML = `There aren't any folders in this directory.`;
        return;
      }

      for(let f of JSON.parse(request.response)) {
        if(f.IsDir === true) {
          dirs++;

          let newNode = document.createElement("li");
          newNode.dataset.url = f.URL;
          newNode.innerHTML = f.Name;
          newNode.setAttribute("aria-selected", false);

          newNode.addEventListener("dblclick", loadNextFolder);
          newNode.addEventListener("click", selectMoveFolder);

          prompt.querySelector("div.file-list ul").appendChild(newNode);
        }
      }

      if(dirs === 0) {
        prompt.querySelector("p").innerHTML = `There aren't any folders in this directory.`;
      }
    }
  }
}

function moveSelected(event) {
  event.preventDefault();

  let promises = [];
  buttons.setLoading("move");

  for(let file of selectedItems) {
    let fileElement = document.getElementById(file),
        destFolder = event.target.querySelector("p code").innerHTML;
    if(event.srcElement.querySelector("li[aria-selected=true]") != null) destFolder = event.srcElement.querySelector("li[aria-selected=true]").innerHTML;
    promises.push(webdav.move(fileElement.dataset.url, "/" + destFolder + "/" + fileElement.querySelector(".name").innerHTML));
  }

  Promise.all(promises)
    .then(() => {
      closePrompt(event);
      buttons.setDone("move");
      listing.reload();
    })
    .catch(e => {
      console.log(e);
    })
}

function moveEvent(event) {
  if(event.currentTarget.classList.contains("disabled"))
    return;

  let request = new XMLHttpRequest();
  request.open("GET", window.location.pathname, true);
  request.setRequestHeader("Accept", "application/json");
  request.send();
  request.onreadystatechange = function () {
    if(request.readyState == 4) {
      if(request.status == 200) {
        let prompt = document.importNode(templates.move.content, true),
          dirs = 0;

        prompt.querySelector("form").addEventListener("submit", moveSelected);
        prompt.querySelector('code').innerHTML = window.location.pathname;

        for(let f of JSON.parse(request.response)) {
          if(f.IsDir === true) {
            dirs++;

            let newNode = document.createElement("li");
            newNode.dataset.url = f.URL;
            newNode.innerHTML = f.Name;
            newNode.setAttribute("aria-selected", false);

            newNode.addEventListener("dblclick", loadNextFolder);
            newNode.addEventListener("click", selectMoveFolder);

            prompt.querySelector("div.file-list ul").appendChild(newNode);
          }
        }

        if(dirs === 0) {
          prompt.querySelector("p").innerHTML = `There aren't any folders in this directory.`;
        }

        document.body.appendChild(prompt);
        document.querySelector(".overlay").classList.add("active");
        document.querySelector(".prompt").classList.add("active");
      }
    }
  }
}

function deleteOnSingleFile() {
  closePrompt(event);
  buttons.setLoading('delete');

  let promises = [];

  webdav.delete(window.location.pathname)
    .then(() => {
      window.location.pathname = removeLastDirectoryPartOf(window.location.pathname);
    })
    .catch(e => {
      buttons.setDone('delete', false);
      console.log(e)
    })
}

function deleteOnListing() {
  closePrompt(event);
  buttons.setLoading('delete');

  let promises = [];

  for(let id of selectedItems) {
    promises.push(webdav.delete(document.getElementById(id).dataset.url));
  }

  Promise.all(promises)
    .then(() => {
      listing.reload();
      buttons.setDone('delete');
    })
    .catch(e => {
      console.log(e);
      buttons.setDone('delete', false);
    })
}

// Handles the delete button event
function deleteEvent(event) {
  let single = false;

  if(!selectedItems.length) {
    selectedItems = ["placeholder"];
    single = true;
  }

  let clone = document.importNode(templates.question.content, true);
  clone.querySelector('h3').innerHTML = 'Delete files';

  if(single) {
    clone.querySelector('form').addEventListener('submit', deleteOnSingleFile);
    clone.querySelector('p').innerHTML = `Are you sure you want to delete this file/folder?`;
  } else {
    clone.querySelector('form').addEventListener('submit', deleteOnListing);
    clone.querySelector('p').innerHTML = `Are you sure you want to delete ${selectedItems.length} file(s)?`;
  }

  clone.querySelector('input').remove();
  clone.querySelector('.ok').innerHTML = 'Delete';

  document.body.appendChild(clone);
  document.querySelector('.overlay').classList.add('active');
  document.querySelector('.prompt').classList.add('active');

  return false;
}

function resetSearchText() {
  let box = document.querySelector('#search > div div');

  if(user.AllowCommands) {
    box.innerHTML = `Search or use one of your supported commands: ${user.Commands.join(", ")}.`;
  } else {
    box.innerHTML = "Type and press enter to search.";
  }
}

function searchEvent(event) {
  if(this.value.length == 0) {
    resetSearchText();
    return;
  }

  let value = this.value,
    search = document.getElementById('search'),
    scrollable = document.querySelector('#search > div'),
    box = document.querySelector('#search > div div'),
    pieces = value.split(' '),
    supported = false;

  user.Commands.forEach(function (cmd) {
    if(cmd == pieces[0]) {
      supported = true;
    }
  });

  if(!supported || !user.AllowCommands) {
    box.innerHTML = "Press enter to search."
  } else {
    box.innerHTML = "Press enter to execute."
  }

  if(event.keyCode == 13) {
    box.innerHTML = '';
    search.classList.add('ongoing');

    let url = window.location.host + window.location.pathname;

    if(document.getElementById("editor")) {
      url = removeLastDirectoryPartOf(url);
    }

    let protocol = ssl ? 'wss:' : 'ws';

    if(supported && user.AllowCommands) {
      let conn = new WebSocket(`${protocol}//${url}?command=true`);

      conn.onopen = function () {
        conn.send(value);
      };

      conn.onmessage = function (event) {
        box.innerHTML = event.data;
        scrollable.scrollTop = scrollable.scrollHeight;
      }

      conn.onclose = function (event) {
        search.classList.remove('ongoing');
        listing.reload();
      }

      return;
    }

    box.innerHTML = '<ul></ul>';

    let ul = box.querySelector('ul'),
      conn = new WebSocket(`${protocol}//${url}?search=true`);

    conn.onopen = function () {
      conn.send(value);
    };

    conn.onmessage = function (event) {
      ul.innerHTML += '<li><a href="' + event.data + '">' + event.data + '</a></li>';
      scrollable.scrollTop = scrollable.scrollHeight;
    }

    conn.onclose = function (event) {
      search.classList.remove('ongoing');
    }
  }
}

function setupSearch() {
  let search = document.getElementById("search"),
    searchInput = search.querySelector("input"),
    searchDiv = search.querySelector("div"),
    hover = false,
    focus = false;

  resetSearchText();

  searchInput.addEventListener('focus', event => {
    focus = true;
    search.classList.add('active');
  });

  searchDiv.addEventListener('mouseover', event => {
    hover = true;
    search.classList.add('active');
  });

  searchInput.addEventListener('blur', event => {
    focus = false;
    if(hover)
      return;
    search.classList.remove('active');
  });

  search.addEventListener('mouseleave', event => {
    hover = false;
    if(focus)
      return;
    search.classList.remove('active');
  });

  search.addEventListener("click", event => {
    search.classList.add("active");
    search.querySelector("input").focus();
  });

  searchInput.addEventListener('keyup', searchEvent);
}

function closeHelp(event) {
  event.preventDefault();

  document.querySelector('.help').classList.remove('active');
  document.querySelector('.overlay').classList.remove('active');
}

function openHelp(event) {
  closePrompt(event);

  document.querySelector('.help').classList.add('active');
  document.querySelector('.overlay').classList.add('active');
}

window.addEventListener('keydown', (event) => {
  if(event.keyCode == 27) {
    if(document.querySelector('.help.active')) {
      closeHelp(event);
    }
  }

  if(event.keyCode == 46) {
    deleteEvent(event);
  }

  if(event.keyCode == 112) {
    event.preventDefault();
    openHelp(event);
  }
});

/* * * * * * * * * * * * * * * *
 *                             *
 *           BOOTSTRAP         *
 *                             *
 * * * * * * * * * * * * * * * */

document.addEventListener("DOMContentLoaded", function (event) {
  overlay = document.querySelector('.overlay');
  clickOverlay = document.querySelector('#click-overlay');

  buttons.logout = document.getElementById("logout");
  buttons.open = document.getElementById("open");
  buttons.delete = document.getElementById("delete");
  buttons.previous = document.getElementById("previous");
  buttons.move = document.getElementById("move");

  // Attach event listeners
  buttons.logout.addEventListener("click", logoutEvent);
  buttons.open.addEventListener("click", openEvent);
  buttons.move.addEventListener("click", moveEvent);

  templates.question = document.querySelector('#question-template');
  templates.info = document.querySelector('#info-template');
  templates.move = document.querySelector("#move-template");

  if(user.AllowEdit) {
    buttons.delete.addEventListener("click", deleteEvent);
  }

  let dropdownButtons = document.querySelectorAll('.action[data-dropdown]')
  Array.from(dropdownButtons).forEach(button => {
    button.addEventListener("click", event => {
      button.querySelector('ul').classList.toggle("active");
      clickOverlay.classList.add('active');

      clickOverlay.addEventListener('click', event => {
        button.querySelector('ul').classList.remove("active");
        clickOverlay.classList.remove('active');
      })
    });
  });

  overlay.addEventListener('click', event => {
    if(document.querySelector('.help.active')) {
      closeHelp(event);
      return;
    }

    closePrompt(event);
  })

  let mainActions = document.getElementById('main-actions');

  document.getElementById('more').addEventListener('click', event => {
    event.preventDefault();
    event.stopPropagation();

    clickOverlay.classList.add('active');
    mainActions.classList.add('active');

    clickOverlay.addEventListener('click', event => {
      mainActions.classList.remove('active');
      clickOverlay.classList.remove('active');
    })
  })

  setupSearch();
  return false;
});
