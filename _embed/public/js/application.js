'use strict';

var tempID = "_fm_internal_temporary_id"
var selectedItems = [];

/* * * * * * * * * * * * * * * *
 *                             *
 *      GENERAL FUNCTIONS      *
 *                             *
 * * * * * * * * * * * * * * * */

// Removes an element, if exists, from an array
Array.prototype.removeElement = function(element) {
    var i = this.indexOf(element);
    if (i != -1) {
        this.splice(i, 1);
    }
}

// Replaces an element inside an array by another
Array.prototype.replaceElement = function(begin, end) {
    var i = this.indexOf(begin);
    if (i != -1) {
        this[i] = end;
    }
}

// Sends a costum event to itself
Document.prototype.sendCostumEvent = function(text) {
    this.dispatchEvent(new CustomEvent(text));
}

// Gets the content of a cookie
Document.prototype.getCookie = function(name) {
    var re = new RegExp("(?:(?:^|.*;\\s*)" + name + "\\s*\\=\\s*([^;]*).*$)|^.*$");
    return document.cookie.replace(re, "$1");
}

// Changes a button to the loading animation
Element.prototype.changeToLoading = function() {
    let element = this;
    let originalText = element.innerHTML;

    element.style.opacity = 0;

    setTimeout(function() {
        element.innerHTML = '<i class="material-icons spin">autorenew</i>';
        element.style.opacity = 1;
    }, 200);

    return originalText;
}

// Changes an element to done animation
Element.prototype.changeToDone = function(error, html) {
    this.style.opacity = 0;

    let thirdStep = () => {
        this.innerHTML = html;
        this.style.opacity = null;

        if (selectedItems.length == 0 && document.getElementById('listing')) {
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

var toWebDavURL = function(url) {
    url = url.replace(baseURL + "/", webdavURL + "/");
    return window.location.origin + url
}

// Handles the open file button event
var openEvent = function(event) {
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

// Handles the delete button event
var deleteEvent = function(event) {
    let single = false;

    if (!selectedItems.length) {
        selectedItems = [window.location.pathname];
        single = true;
    }

    if (selectedItems.length) {
        Array.from(selectedItems).forEach(link => {
            let html = document.getElementById("delete").changeToLoading();
            let request = new XMLHttpRequest();

            request.open('DELETE', toWebDavURL(link));

            request.send();
            request.onreadystatechange = function() {
                if (request.readyState == 4) {
                    if (request.status == 204) {
                        if (single) {
                            window.location.pathname = RemoveLastDirectoryPartOf(window.location.pathname);
                        } else {
                            document.getElementById(link).remove();
                            selectedItems.removeElement(link);
                        }
                    }
                    document.getElementById('delete').changeToDone((request.status != 204), html);
                }
            }
        });

        return false;
    }

    return false;
}

// Prevent Default event
var preventDefault = function(event) {
    event.preventDefault();
}

// Remove the last directory of an url
var RemoveLastDirectoryPartOf = function(url) {
    var arr = url.split('/');
    arr.pop();
    return (arr.join('/'));
}

/* * * * * * * * * * * * * * * *
 *                             *
 *  LISTING SPECIFIC FUNCTIONS  *
 *                             *
 * * * * * * * * * * * * * * * */

var reloadListing = function(callback) {
    let request = new XMLHttpRequest();
    request.open('GET', window.location);
    request.setRequestHeader('Minimal', 'true');
    request.send();
    request.onreadystatechange = function() {
        if (request.readyState == 4) {
            if (request.status == 200) {
                document.querySelector('body main').innerHTML = request.responseText;
                // Handle date times
                let timeList = document.getElementsByTagName("time");
                Array.from(timeList).forEach(localizeDatetime);

                // Add action to checkboxes
                let checkboxes = document.getElementsByClassName('checkbox');
                Array.from(checkboxes).forEach(link => {
                    link.addEventListener('click', itemClickEvent);
                });

                addNewDirEvents();

                if (typeof callback == 'function') {
                    callback();
                }
            }
        }
    }
}

// Rename file event
var renameEvent = function(event) {
    if (this.classList.contains('disabled') || !selectedItems.length) {
        return false;
    }

    // This mustn't happen
    if (selectedItems.length > 1) {
        alert("Something went wrong. Please refresh the page.");
        location.refresh();
    }

    let link = selectedItems[0];
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
            let newLink = toWebDavURL(link).replace(name, newName)
            let html = document.getElementById('rename').changeToLoading();
            let request = new XMLHttpRequest();
            request.open('MOVE', toWebDavURL(link));
            request.setRequestHeader('Destination', newLink);

            request.send();
            request.onreadystatechange = function() {
                // TODO: redirect if it's moved to another folder

                if (request.readyState == 4) {
                    if (request.status != 201 && request.status != 204) {
                        span.innerHTML = name;
                    } else {
                        let newLink = encodeURI(link.replace(name, newName));
                        console.log(request.body)
                        reloadListing(() => {
                            let newLink = encodeURI(link.replace(name, newName));
                            selectedItems = [newLink];
                            document.getElementById(newLink).classList.add("selected")
                            var event = new CustomEvent('changed-selected');
                            document.dispatchEvent(event);
                        });
                    }

                    document.getElementById('rename').changeToDone((request.status != 201 && request.status != 204), html);
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

    return false;
}

// Upload files
var handleFiles = function(files, base) {
    let button = document.getElementById("upload");
    let html = button.changeToLoading();

    for (let i = 0; i < files.length; i++) {
        let request = new XMLHttpRequest();
        request.open('PUT', toWebDavURL(window.location.pathname + base + files[i].name));

        request.send(files[i]);
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                if (request.status == 201) {
                    reloadListing();
                }

                button.changeToDone((request.status != 201), html);
            }
        }
    }

    return false;
}

// Handles the back button event
var backEvent = function(event) {
    var items = document.getElementsByClassName('item');
    Array.from(items).forEach(link => {
        link.classList.remove('selected');
    });
    selectedItems = [];

    var event = new CustomEvent('changed-selected');
    document.dispatchEvent(event);
    return false;
}

// Handles the click event
var itemClickEvent = function(event) {
    var url = this.dataset.href;
    var el = document.getElementById(url);

    if (selectedItems.length != 0) event.preventDefault();
    if (selectedItems.indexOf(url) == -1) {
        el.classList.add('selected');
        selectedItems.push(url);
    } else {
        el.classList.remove('selected');
        selectedItems.removeElement(url);
    }

    var event = new CustomEvent('changed-selected');
    document.dispatchEvent(event);
    return false;
}

// Handles the datetimes present on the document
var localizeDatetime = function(e, index, ar) {
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

// Toggles the view mode
var viewEvent = function(event) {
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

// Handles the view mode change
var handleViewType = function(viewList) {
    let listing = document.getElementById('listing');
    let button = document.getElementById('view');

    if (viewList == "true") {
        listing.classList.add('list');
        button.innerHTML = '<i class="material-icons" title="Switch View">view_module</i> <span>Switch view</span>';
        return false;
    }

    button.innerHTML = '<i class="material-icons" title="Switch View">view_list</i> <span>Switch view</span>';
    listing.classList.remove('list');
    return false;
}

var addNewDirEvents = function() {
    document.getElementById('new').addEventListener('click', event => {
        let newdir = document.getElementById('newdir');
        newdir.classList.add('enabled');
        newdir.focus();
    });

    document.getElementById('newdir').addEventListener('blur', event => {
        document.getElementById('newdir').classList.remove('enabled');
    });

    document.getElementById('newdir').addEventListener('keydown', newDirEvent);
}

// Handles the new directory event
var newDirEvent = function(event) {
    // TODO: create new dir button and new file button
    if (event.keyCode == 27) {
        document.getElementById('newdir').classList.toggle('enabled');
        setTimeout(() => {
            document.getElementById('newdir').value = '';
        }, 200);
    }

    if (event.keyCode == 13) {
        event.preventDefault();

        let button = document.getElementById('new');
        let html = button.changeToLoading();
        let request = new XMLHttpRequest();
        let name = document.getElementById('newdir').value;

        request.open((name.endsWith("/") ? "MKCOL" : "PUT"), toWebDavURL(window.location.pathname + name));

        request.send();
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                button.changeToDone((request.status != 201), html);
                reloadListing(() => {
                    addNewDirEvents();
                });
            }
        }
    }

    return false;
}

// Handles the event when there is change on selected elements
document.addEventListener("changed-selected", function(event) {
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

        redefineDownloadURLs();

        return false;
    }

    toolbar.classList.remove("enabled");
    return false;
});

var redefineDownloadURLs = function() {
    let files = "";

    for (let i = 0; i < selectedItems.length; i++) {
        files += selectedItems[i].replace(window.location.pathname, "") + ",";
    }

    files = files.substring(0, files.length - 1);
    files = encodeURIComponent(files);

    let links = document.querySelectorAll("#download ul a");
    Array.from(links).forEach(link => {
        link.href = "?download=" + link.dataset.format + "&files=" + files;
    });
}

var searchEvent = function(event) {
    let value = this.value;
    let search = document.getElementById('search');
    let scrollable = document.querySelector('#search > div');
    let box = document.querySelector('#search > div div');

    if (value.length == 0) {
        box.innerHTML = "Search or use one of your supported commands: " + user.Commands.join(", ") + ".";
        return;
    }

    let pieces = value.split(' ');
    let supported = false;

    user.Commands.forEach(function(cmd) {
        if (cmd == pieces[0]) {
            supported = true;
        }
    });

    if (!supported) {
        box.innerHTML = "Press enter to search."
    } else {
        box.innerHTML = "Press enter to execute."
    }

    if (event.keyCode == 13) {
        box.innerHTML = '';
        search.classList.add('ongoing');

        if (supported) {
            var conn = new WebSocket('ws://' + window.location.host + window.location.pathname + '?command=true');
            conn.onopen = function() {
                conn.send(value);
            };

            conn.onmessage = function(event) {
                box.innerHTML = event.data;
                scrollable.scrollTop = scrollable.scrollHeight;
            }

            conn.onclose = function(event) {
                search.classList.remove('ongoing');
                reloadListing();
            }
        } else {
            box.innerHTML = '<ul></ul>';
            let ul = box.querySelector('ul');

            var conn = new WebSocket('ws://' + window.location.host + window.location.pathname + '?search=true');
            conn.onopen = function() {
                conn.send(value);
            };

            conn.onmessage = function(event) {
                ul.innerHTML += '<li><a href="' + event.data + '">' + event.data + '</a></li>';
                scrollable.scrollTop = scrollable.scrollHeight;
            }

            conn.onclose = function(event) {
                search.classList.remove('ongoing');
            }
        }
    }
}

document.addEventListener('listing', event => {
    // Handle date times
    let timeList = document.getElementsByTagName("time");
    Array.from(timeList).forEach(localizeDatetime);

    // Handles the current view mode and adds the event to the button
    handleViewType(document.getCookie("view-list"));
    document.getElementById("view").addEventListener("click", viewEvent);

    // Add event to items
    let checkboxes = document.getElementsByClassName('checkbox');
    Array.from(checkboxes).forEach(link => {
        link.addEventListener('click', itemClickEvent);
    });

    // Add event to back button and executes back event on ESC
    document.getElementById("back").addEventListener("click", backEvent)
    document.addEventListener('keydown', (event) => {
        if (event.keyCode == 27) {
            backEvent(event);
        }
    });

    if (user.AllowCommands) {
        let hover = false,
            focus = false;

        document.querySelector('#search input').addEventListener('focus', event => {
            focus = true;
            document.getElementById('search').classList.add('active');
        });

        document.querySelector('#search div').addEventListener('mouseover', event => {
            hover = true;
            document.getElementById('search').classList.add('active');
        });

        document.querySelector('#search input').addEventListener('blur', event => {
            focus = false;
            if (hover) return;
            document.getElementById('search').classList.remove('active');
        });

        document.querySelector('#search').addEventListener('mouseleave', event => {
            hover = false;
            if (focus) return;
            document.getElementById('search').classList.remove('active');
        });

        document.querySelector('#search > div div').innerHTML = "Search or use one of your supported commands: " + user.Commands.join(", ") + ".";
        document.querySelector('#search input').addEventListener('keyup', searchEvent);
    }

    if (user.AllowEdit) {
        // Enables rename button
        document.getElementById("rename").addEventListener("click", renameEvent);
    }

    if (user.AllowNew) {
        // Enables upload button
        document.getElementById("upload").addEventListener("click", (event) => {
            document.getElementById("upload-input").click();
        });

        addNewDirEvents();

        // Drag and Drop
        let items = document.getElementsByClassName('item');
        document.addEventListener("dragover", function(event) {
            event.preventDefault();
        }, false);

        document.addEventListener("dragenter", (event) => {
            Array.from(items).forEach(file => {
                file.style.opacity = 0.5;
            });
        }, false);

        document.addEventListener("dragend", (event) => {
            Array.from(items).forEach(file => {
                file.style.opacity = 1;
            });
        }, false);

        document.addEventListener("drop", function(event) {
            event.preventDefault();
            var dt = event.dataTransfer;
            var files = dt.files;

            let el = event.target;

            for (let i = 0; i < 5; i++) {
                if (el != null && !el.classList.contains('item')) {
                    el = el.parentElement;
                }
            }

            if (files.length > 0) {
                if (el != null && el.classList.contains('item') && el.dataset.dir == "true") {
                    handleFiles(files, el.querySelector('.name').innerHTML + "/");
                    return;
                }

                handleFiles(files, "");
            } else {
                Array.from(items).forEach(file => {
                    file.style.opacity = 1;
                });
            }

        }, false);
    }
});

function itemDragStart(event) {
    let el = event.target;

    for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
            el = el.parentElement;
        }
    }

    event.dataTransfer.setData("id", el.id);
    event.dataTransfer.setData("name", el.querySelector('.name').innerHTML);
}

function itemDragOver(event) {
    event.preventDefault();
    let el = event.target;

    for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
            el = el.parentElement;
        }
    }

    el.style.opacity = 1;
}

function itemDrop(e) {
    e.preventDefault();

    let el = e.target, id = e.dataTransfer.getData("id"), name = e.dataTransfer.getData("name");
    if (id == "" || name == "") return;

    for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
            el = el.parentElement;
        }
    }

    if (el.id === id) return;

    let oldLink = toWebDavURL(id);
    let newLink = toWebDavURL(el.id + name);

    let request = new XMLHttpRequest();
    request.open('MOVE', oldLink);
    request.setRequestHeader('Destination', newLink);
    request.send();
    request.onreadystatechange = function() {
        if (request.readyState == 4) {
            if (request.status == 201 || request.status == 204) {
                reloadListing();
            }
        }
    }
}


/* * * * * * * * * * * * * * * *
 *                             *
 *  EDITOR SPECIFIC FUNCTIONS  *
 *                             *
 * * * * * * * * * * * * * * * */

// auto grow textareas
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

var deleteFrontMatterItem = function(event) {
    event.preventDefault();
    document.getElementById(this.dataset.delete).remove();
}

var addFrontMatterItem = function(event) {
    event.preventDefault();

    let temp = document.getElementById(tempID)
    if (temp) {
        temp.remove();
    }

    let block = this.parentNode;
    let type = block.dataset.type;
    let id = block.id;

    // If the block is an array
    if (type === "array") {
        let fieldID = id + "[]"
        let input = fieldID
        let count = block.querySelectorAll('.group > div').length
        input = input.replace(/\[/, '\\[');
        input = input.replace(/\]/, '\\]');

        let fieldsets = block.getElementsByTagName("fieldset");

        if (fieldsets.length > 0) {
            let newtype = fieldsets[0].dataset.type;
            let bid = id + "[" + fieldsets.length + "]";
            let name = fieldsets.length;

            let template = `<fieldset id="${bid}" data-type="${newtype}">
              <h3>${name}</h3>
              <div class="action add">
                  <i class="material-icons">add</i>
              </div>
              <div class="action delete" data-delete="${bid}">
                  <i class="material-icons">close</i>
              </div>
             <div class="group">
             </div>
            </fieldset>`;

            block.insertAdjacentHTML('beforeend', template);

            document.querySelector(`div[data-delete="${bid}"]`).addEventListener('click', deleteFrontMatterItem);
            document.getElementById(bid).querySelector('.action.add').addEventListener('click', addFrontMatterItem);
        } else {
            block.querySelector('.group').insertAdjacentHTML('beforeend', `<div id="${fieldID}-${count}" data-type="array-item">
                <input name="${fieldID}" id="${fieldID}" type="text" data-parent-type="array"></input>
                <div class="action delete"  data-delete="${fieldID}-${count}">
                    <i class="material-icons">close</i>
                </div>
            </div>`);

            document.getElementById(`${fieldID}-${count}`).querySelector('input').focus();
            document.querySelector(`div[data-delete="${fieldID}-${count}"]`).addEventListener('click', deleteFrontMatterItem);
        }


    }

    if (type == "object" || type == "parent") {
        let template = `<div class="group temp" id="${tempID}">
        <div class="block" id="${tempID}">
            <label>Write the field name and then press enter. If you want to create an array or an object, end the name with ":array" or ":object".</label>
            <input name="${tempID}" type="text" placeholder="Write the field name and press enter.."></input>
        </div></div>`;

        if (type == "parent") {
            document.querySelector('div.button.add').insertAdjacentHTML('beforebegin', template);
        } else {
            block.querySelector('.delete').insertAdjacentHTML('afterend', template);
        }

        let temp = document.getElementById(tempID);
        let input = temp.querySelector('input');
        input.focus();
        input.addEventListener('keydown', (event) => {
            if (event.keyCode == 27) {
                event.preventDefault();
                temp.remove();
            }

            if (event.keyCode == 13) {
                event.preventDefault();

                let value = input.value;
                if (value === '') {
                    temp.remove();
                    return true;
                }

                let name = value.substring(0, value.lastIndexOf(':'));
                let newtype = value.substring(value.lastIndexOf(':') + 1, value.length);
                if (newtype !== "" && newtype !== "array" && newtype !== "object") {
                    name = value;
                }

                name = name.replace(' ', '_');

                let bid = name;
                if (id != '') {
                    bid = id + "." + bid;
                }

                temp.remove();

                switch (newtype) {
                    case "array":
                    case "object":
                        let template = `<fieldset id="${bid}" data-type="${newtype}">
                          <h3>${name}</h3>
                          <div class="action add">
                              <i class="material-icons">add</i>
                          </div>
                          <div class="action delete" data-delete="${bid}">
                              <i class="material-icons">close</i>
                          </div>
                         <div class="group">
                         </div>
                        </fieldset>`;

                        if (type == "parent") {
                            document.querySelector('div.button.add').insertAdjacentHTML('beforebegin', template);
                        } else {
                            block.insertAdjacentHTML('beforeend', template);
                        }

                        document.querySelector(`div[data-delete="${bid}"]`).addEventListener('click', deleteFrontMatterItem);
                        document.getElementById(bid).querySelector('.action.add').addEventListener('click', addFrontMatterItem);
                        break;
                    default:
                        let group = block.querySelector('.group');

                        if (group == null) {
                            block.insertAdjacentHTML('afterbegin', '<div class="group"></div>');
                            group = block.querySelector('.group');
                        }

                        group.insertAdjacentHTML('beforeend', `<div class="block" id="block-${bid}" data-content="${bid}">
                          <label for="${bid}">${name}</label>
                          <input name="${bid}" id="${bid}" type="text" data-parent-type="object"></input>
                          <div class="action delete" data-delete="block-${bid}">
                              <i class="material-icons">close</i>
                          </div>
                        </div>`);

                        document.getElementById(bid).focus();
                        document.querySelector(`div[data-delete="block-${bid}"]`).addEventListener('click', deleteFrontMatterItem);
                }
            }
        });
    }

    return false;
}

document.addEventListener("editor", (event) => {
    textareaAutoGrow();

    let container = document.getElementById('editor');
    let button = document.querySelector('#submit span:first-child');
    let kind = container.dataset.kind;

    if (kind != 'frontmatter-only') {
        let editor = document.getElementById('editor-source');
        let mode = editor.dataset.mode;
        let textarea = document.querySelector('textarea[name="content"]');
        let aceEditor = ace.edit('editor-source');
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

    let deleteFrontMatterItemButtons = document.getElementsByClassName('delete');
    Array.from(deleteFrontMatterItemButtons).forEach(button => {
        button.addEventListener('click', deleteFrontMatterItem);
    });

    let addFrontMatterItemButtons = document.getElementsByClassName('add');
    Array.from(addFrontMatterItemButtons).forEach(button => {
        button.addEventListener('click', addFrontMatterItem);
    });

    let saveContent = function() {
        let data = form2js(document.querySelector('form'));

        if (typeof data.content === "undefined" && kind != 'frontmatter-only') {
            data.content = "";
        }

        if (typeof data.content === "number") {
            data.content = data.content.toString();
        }

        let html = button.changeToLoading();
        let request = new XMLHttpRequest();
        request.open("PUT", toWebDavURL(window.location.pathname));
        request.setRequestHeader('Kind', kind);

        request.send(JSON.stringify(data));
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                button.changeToDone((request.status != 201), html);
            }
        }
    }

    document.querySelector('form').addEventListener('submit', (event) => {
        event.preventDefault();
        saveContent();
    });

    window.addEventListener('keydown', (event) => {
        if (event.ctrlKey || event.metaKey) {
            switch (String.fromCharCode(event.which).toLowerCase()) {
                case 's':
                    event.preventDefault();
                    saveContent();
                    break;
            }
        }
    });

    return false;
});

/* * * * * * * * * * * * * * * *
 *                             *
 *           BOOTSTRAP         *
 *                             *
 * * * * * * * * * * * * * * * */

document.addEventListener("DOMContentLoaded", function(event) {
    // Add event to logout button
    document.getElementById("logout").addEventListener("click", event => {
        let request = new XMLHttpRequest();
        request.open('GET', window.location.pathname, true, "username", "password");
        request.send();
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                window.location = "/";
            }
        }
    });

    // Enables open, delete and download buttons
    document.getElementById("open").addEventListener("click", openEvent);

    if (user.AllowEdit) {
        document.getElementById("delete").addEventListener("click", deleteEvent);
    }

    document.getElementById("open-nav").addEventListener("click", event => {
        document.querySelector("header > div:nth-child(2)").classList.toggle("active");
    });
    document.getElementById("overlay").addEventListener("click", event => {
        document.querySelector("header > div:nth-child(2)").classList.toggle("active");
    });

    if (document.getElementById('listing')) {
        document.sendCostumEvent('listing');
    }

    if (document.getElementById('editor')) {
        document.sendCostumEvent('editor');
    }

    return false;
});
