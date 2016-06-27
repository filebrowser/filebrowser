"use strict";

var selectedItems = [];

// Array prototype function to remove an element
Array.prototype.removeElement = function (element) {
 var i = this.indexOf(element);
 if (i != -1) {
  this.splice(i, 1);
 }
}

// Array prototype function to replace an element
Array.prototype.replaceElement = function (oldEl, newEl) {
 var i = this.indexOf(oldEl);
 if (i != -1) {
  this[i] = newEl;
 }
}

// Document prototype function to send a costum event to itself
Document.prototype.sendCostumEvent = function (text) {
 document.dispatchEvent(new CustomEvent(text));
}

// Document prototype to get a cookie content
Document.prototype.getCookie = function (name) {
 var re = new RegExp("(?:(?:^|.*;\\s*)" + name + "\\s*\\=\\s*([^;]*).*$)|^.*$");
 return document.cookie.replace(re, "$1");
}

// Changes a button to the loading animation
Element.prototype.changeToLoading = function () {
 let element = this;
 let originalText = element.innerHTML;

 element.style.opacity = 0;

 setTimeout(function () {
  element.innerHTML = '<i class="material-icons spin">autorenew</i>';
  element.style.opacity = 1;
 }, 200);

 return originalText;
}

// Changes an element to done animation
Element.prototype.changeToDone = function (error, html) {
 this.style.opacity = 0;

 let thirdStep = () => {
  this.innerHTML = html;
  this.style.opacity = 1;

  if (selectedItems.length == 0) {
   document.sendCostumEvent('changed-selected');
  }
 }

 let secondStep = () => {
  this.style.opacity = 0;
  setTimeout(thirdStep, 200);
 }

 let firstStep = () => {
  if (error) {
   this.innerHTML = '<i class="material-icons">close</i>';
  } else {
   this.innerHTML = '<i class="material-icons">done</i>';
  }

  this.style.opacity = 1;

  setTimeout(secondStep, 1000);
 }

 setTimeout(firstStep, 200);
 return false;
}

// Event for toggling the mode of view
var viewEvent = function (event) {
 let cookie = document.getCookie('view-list');
 let listing = document.getElementById('listing');

 if (cookie != 'true') {
  document.cookie = 'view-list=true';
 } else {
  document.cookie = 'view-list=false';
 }

 handleViewType(document.getCookie('view-list'));
 return false;
}

// Handles the view type change
var handleViewType = function (viewList) {
 let listing = document.getElementById('listing');
 let button = document.getElementById('view');

 if (viewList == "true") {
  listing.classList.add('list');
  button.innerHTML = '<i class="material-icons">view_module</i>';
  return false;
 }

 button.innerHTML = '<i class="material-icons">view_list</i>';
 listing.classList.remove('list');
 return false;
}

// Handles the open file button event
var openEvent = function (event) {
 if (this.classList.contains('disabled')) {
  return false;
 }
 if (selectedItems.length) {
  window.open(selectedItems[0] + '?raw=true');
  return false;
 }

 window.open(window.location + '?raw=true');
 return false;
}

// Handles the back button event
var backEvent = function (event) {
 var items = document.getElementsByClassName('item');
 Array.from(items).forEach(link => {
  link.classList.remove('selected');
 });
 selectedItems = [];

 var event = new CustomEvent('changed-selected');
 document.dispatchEvent(event);
 return false;
}

// Handles the delete button event
var deleteEvent = function (event) {
 if (selectedItems.length) {
  Array.from(selectedItems).forEach(link => {
   let html = document.getElementById("delete").changeToLoading();
   let request = new XMLHttpRequest();

   request.open('DELETE', link);
   request.send();
   request.onreadystatechange = function () {
    if (request.readyState == 4) {
     if (request.status == 200) {
      document.getElementById(link).remove();
      console.log(selectedItems);
      selectedItems.removeElement(link);
     }

     document.getElementById('delete').changeToDone((request.status != 200), html);
    }
   }
  });

  return false;
 }

 let request = new XMLHttpRequest();
 request.open('DELETE', window.location);
 request.send();
 request.onreadystatechange = function () {
  if (request.readyState == 4) {
   if (request.status == 200) {
    window.location.pathname = RemoveLastDirectoryPartOf(window.location.pathname);
   }

   document.getElementById('delete').changeToDone((request.status != 200), html);
  }
 }

 return false;
}

// Prevent Default event
var preventDefault = function (event) {
 event.preventDefault();
}

// Rename file event
var renameEvent = function (event) {
 if (this.classList.contains('disabled')) {
  return false;
 }
 if (selectedItems.length) {
  Array.from(selectedItems).forEach(link => {
   let item = document.getElementById(link);
   let span = item.getElementsByTagName('span')[0];
   let name = span.innerHTML;

   item.addEventListener('click', preventDefault);
   item.removeEventListener('click', itemClickEvent);
   span.setAttribute('contenteditable', 'true');
   span.focus();

   let keyDownEvent = (event) => {
    if (event.keyCode == 13) {
     let newName = span.innerHTML;
     let html = document.getElementById('rename').changeToLoading();
     let request = new XMLHttpRequest();
     request.open('PATCH', link);
     request.setRequestHeader('Rename-To', newName);
     request.send();
     request.onreadystatechange = function () {
      if (request.readyState == 4) {
       if (request.status != 200) {
        span.innerHTML = name;
       } else {
        let newLink = link.replace(name, newName);
        item.id = newLink;
        selectedItems.replaceElement(link, newLink);
        span.innerHTML = newName;
       }

       document.getElementById('rename').changeToDone((request.status != 200), html);
      }
     }
    }

    if (event.KeyCode == 27) {
     span.innerHTML = name;
    }

    if (event.keyCode == 13 || event.keyCode == 27) {
     span.setAttribute('contenteditable', 'false');
     span.removeEventListener('keydown', keyDownEvent);
     item.removeEventListener('click', preventDefault);
     item.addEventListener('click', itemClickEvent);
     event.preventDefault();
    }

    return false;
   }

   span.addEventListener('keydown', keyDownEvent);
   span.addEventListener('blur', (event) => {
    span.innerHTML = name;
    span.setAttribute('contenteditable', 'false');
    span.removeEventListener('keydown', keyDownEvent);
    item.removeEventListener('click', preventDefault);
   });
  });

  return false;
 }

 return false;
}

// Download file event
var downloadEvent = function (event) {
 if (this.classList.contains('disabled')) {
  return false;
 }
 if (selectedItems.length) {
  Array.from(selectedItems).forEach(item => {
   window.open(item + "?download=true");
  });
  return false;
 }

 window.open(window.location + "?download=true");
 return false;
}

var handleFiles = function (files) {
 let button = document.getElementById("upload");
 let html = button.changeToLoading();
 let data = new FormData();

 for (let i = 0; i < files.length; i++) {
  data.append(files[i].name, files[i]);
 }

 let request = new XMLHttpRequest();
 request.open('POST', window.location.pathname);
 request.setRequestHeader("Upload", "true");
 request.send(data);
 request.onreadystatechange = function () {
  if (request.readyState == 4) {
   if (request.status == 200) {
    location.reload();
   }

   button.changeToDone((request.status != 200), html);
  }
 }

 return false;
}

var RemoveLastDirectoryPartOf = function (url) {
 var arr = url.split('/');
 arr.pop();
 return (arr.join('/'));
}

document.addEventListener("changed-selected", function (event) {
 var toolbar = document.getElementById("toolbar");
 var selectedNumber = selectedItems.length;

 document.getElementById("selected-number").innerHTML = selectedNumber;

 if (selectedNumber) {
  toolbar.classList.add("enabled");

  if (selectedNumber > 1) {
   document.getElementById("open").classList.add("disabled");
   document.getElementById("rename").classList.add("disabled");
  }

  if (selectedNumber == 1) {
   document.getElementById("open").classList.remove("disabled");
   document.getElementById("rename").classList.remove("disabled");
  }

  return false;
 }

 toolbar.classList.remove("enabled");
 return false;
});

var itemClickEvent = function (event) {
 var url = this.getElementsByTagName('a')[0].getAttribute('href');

 if (selectedItems.length != 0) event.preventDefault();
 if (selectedItems.indexOf(url) == -1) {
  this.classList.add('selected');
  selectedItems.push(url);
 } else {
  this.classList.remove('selected');
  selectedItems.removeElement(url);
 }

 var event = new CustomEvent('changed-selected');
 document.dispatchEvent(event);
 return false;
}

var localizeDatetime = function (e, index, ar) {
 if (e.textContent === undefined) {
  return;
 }
 var d = new Date(e.getAttribute('datetime'));
 if (isNaN(d)) {
  d = new Date(e.textContent);
  if (isNaN(d)) {
   return;
  }
 }
 e.textContent = d.toLocaleString();
}

document.addEventListener("DOMContentLoaded", function (event) {

 document.getElementById("logout").addEventListener("click", event => {
  let request = new XMLHttpRequest();
  request.open('GET', window.location.pathname, true, "username", "password");
  request.send();
  request.onreadystatechange = function () {
   if (request.readyState == 4) {
    window.location = "/";
   }
  }
 });

 var timeList = Array.prototype.slice.call(document.getElementsByTagName("time"));
 timeList.forEach(localizeDatetime);

 var items = document.getElementsByClassName('item');
 Array.from(items).forEach(link => {
  link.addEventListener('click', itemClickEvent);
 });

 document.getElementById("open").addEventListener("click", openEvent);
 if (document.getElementById("back")) {
  document.getElementById("back").addEventListener("click", backEvent)
 };
 if (document.getElementById("view")) {
  handleViewType(document.getCookie("view-list"));
  document.getElementById("view").addEventListener("click", viewEvent)
 };
 if (document.getElementById("upload")) {
  document.getElementById("upload").addEventListener("click", (event) => {
   document.getElementById("upload-input").click();
  });
 }
 document.getElementById("delete").addEventListener("click", deleteEvent);
 document.getElementById("download").addEventListener("click", downloadEvent);

 let rename = document.getElementById("rename");
 if (rename) {
  rename.addEventListener("click", renameEvent);
 }

 if (document.getElementById("listing")) {
  document.addEventListener("dragover", function (event) {
   event.preventDefault();
  }, false);

  document.addEventListener("dragover", (event) => {
   Array.from(items).forEach(file => {
    file.style.opacity = 0.5;
   });
  }, false);

  document.addEventListener("dragleave", (event) => {
   Array.from(items).forEach(file => {
    file.style.opacity = 1;
   });
  }, false);

  document.addEventListener("drop", function (event) {
   event.preventDefault();
   var dt = event.dataTransfer;
   var files = dt.files;

   handleFiles(files);
  }, false);
 }

 if (document.getElementById('editor')) {
  handleEditorPage();
 }

 textareaAutoGrow();

 return false;
});

var textareaAutoGrow = function() {
 let autogrow = function() {
   this.style.height = '5px';
   this.style.height = this.scrollHeight + 'px';
 }

 let textareas = document.getElementsByTagName('textarea');

 let addAutoGrow = () => {
     Array.from(textareas).forEach(textarea => {
        autogrow.bind(textarea)();
        textarea.addEventListener('keyup', autogrow);
     });
 }

 addAutoGrow();
 window.addEventListener('resize', addAutoGrow)
}

var handleEditorPage = function () {
    let container = document.getElementById('editor');
    let kind = container.dataset.kind;

    if (kind != 'frontmatter-only') {
        let editor = document.getElementById('editor-source');
        let mode = editor.dataset.mode;
        let textarea = document.querySelector('textarea[name="content"]');
        let aceEditor =  ace.edit('editor-source');
        aceEditor.getSession().setMode("ace/mode/" + mode);
        aceEditor.getSession().setValue(textarea.value);
        aceEditor.getSession().on('change', function() {
          textarea.value = aceEditor.getSession().getValue();
        });
        aceEditor.setOptions({
          wrap: true,
          maxLines: Infinity,
          theme: "ace/theme/github",
          showPrintMargin: false,
          fontSize: "1em",
          minLines: 20
        });
    }




 return false;
}
