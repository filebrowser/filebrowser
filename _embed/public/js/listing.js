'use strict';

var reloadListing = function(callback) {
    let request = new XMLHttpRequest();
    request.open('GET', window.location);
    request.setRequestHeader('Minimal', 'true');
    request.send();
    request.onreadystatechange = function() {
        if (request.readyState == 4) {
            if (request.status == 200) {
                document.querySelector('body main').innerHTML = request.responseText;

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

    let item = document.getElementById(selectedItems[0]),
        link = item.dataset.url,
        span = item.getElementsByTagName('span')[0],
        name = span.innerHTML;

    span.setAttribute('contenteditable', 'true');
    span.focus();

    let keyDownEvent = (event) => {
        if (event.keyCode == 13) {
            let newName = span.innerHTML,
                newLink = removeLastDirectoryPartOf(toWebDavURL(link)) + "/" + newName,
                html = document.getElementById('rename').changeToLoading(),
                request = new XMLHttpRequest();

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
                            newName = btoa(newName);
                            selectedItems = [newName];
                            document.getElementById(newName).setAttribute("aria-selected", true);
                            document.sendCostumEvent('changed-selected');
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
    let button = document.getElementById("upload"),
        html = button.changeToLoading();

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

function unselectAll() {
    var items = document.getElementsByClassName('item');
    Array.from(items).forEach(link => {
        link.setAttribute("aria-selected", false);
    });
    
    selectedItems = [];
    
    document.sendCostumEvent('changed-selected');
    return false;
}

// Toggles the view mode
var viewEvent = function(event) {
    let cookie = document.getCookie('view-list'),
        listing = document.getElementById('listing');

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
    let listing = document.getElementById('listing'),
        button = document.getElementById('view');

    if (viewList == "true") {
        listing.classList.add('list');
        button.innerHTML = '<i class="material-icons" title="Switch View">view_module</i> <span>Switch view</span>';
        return false;
    }

    button.innerHTML = '<i class="material-icons" title="Switch View">view_list</i> <span>Switch view</span>';
    listing.classList.remove('list');
    return false;
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

        let button = document.getElementById('new'),
            html = button.changeToLoading(),
            request = new XMLHttpRequest(),
            name = document.getElementById('newdir').value;

        request.open((name.endsWith("/") ? "MKCOL" : "PUT"), toWebDavURL(window.location.pathname + name));

        request.send();
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                button.changeToDone((request.status != 201), html);
                reloadListing();
            }
        }
    }

    return false;
}

// Handles the event when there is change on selected elements
document.addEventListener("changed-selected", function(event) {
    redefineDownloadURLs();

    let selectedNumber = selectedItems.length,
        fileAction = document.getElementById("file-only");

    if (selectedNumber) {
        fileAction.classList.remove("disabled");

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

    fileAction.classList.add("disabled");
    return false;
});

var redefineDownloadURLs = function() {
    let files = "";

    for (let i = 0; i < selectedItems.length; i++) {
        let url = document.getElementById(selectedItems[i]).dataset.url;
        files += url.replace(window.location.pathname, "") + ",";
    }

    files = files.substring(0, files.length - 1);
    files = encodeURIComponent(files);

    let links = document.querySelectorAll("#download ul a");
    Array.from(links).forEach(link => {
        link.href = "?download=" + link.dataset.format + "&files=" + files;
    });
}


document.addEventListener('DOMContentLoaded', event => {
    // Handles the current view mode and adds the event to the button
    handleViewType(document.getCookie("view-list"));
    document.getElementById("view").addEventListener("click", viewEvent);

    let updateColumns = () => {
        let columns = Math.floor(document.getElementById('listing').offsetWidth / 300),
            itens = getCSSRule('#listing .item');

        itens.style.width = `calc(${100/columns}% - 1em)`;
    }

    updateColumns();
    window.addEventListener("resize", () => {
        updateColumns();
    });

    // Add event to back button and executes back event on ESC
    document.addEventListener('keydown', (event) => {
        if (event.keyCode == 27) {
            unselectAll();
        }
    });

    if (user.AllowEdit) {
        // Enables rename button
        document.getElementById("rename").addEventListener("click", renameEvent);
    }

    if (user.AllowNew) {
        // Enables upload button
        document.getElementById("upload").addEventListener("click", (event) => {
            document.getElementById("upload-input").click();
        });

        document.getElementById('new').addEventListener('click', event => {
            let newdir = document.getElementById('newdir');
            newdir.classList.add('enabled');
            newdir.focus();
        });
    
        document.getElementById('newdir').addEventListener('blur', event => {
            document.getElementById('newdir').classList.remove('enabled');
        });
    
        document.getElementById('newdir').addEventListener('keydown', newDirEvent);

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

    let el = e.target,
        id = e.dataTransfer.getData("id"),
        name = e.dataTransfer.getData("name");

    if (id == "" || name == "") return;

    for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
            el = el.parentElement;
        }
    }

    if (el.id === id) return;

    let oldLink = toWebDavURL(document.getElementById(id).dataset.url),
        newLink = toWebDavURL(el.dataset.url + name),
        request = new XMLHttpRequest();

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

function openItemEvent(event) {
    window.location = event.currentTarget.dataset.url;
}

function selectItemEvent(event) {
    let el = event.currentTarget;

    if (selectedItems.length != 0) event.preventDefault();
    if (selectedItems.indexOf(el.id) == -1) {
        if (!event.ctrlKey) unselectAll();

        el.setAttribute("aria-selected", true);
        selectedItems.push(el.id);
    } else {
        el.setAttribute("aria-selected", false);
        selectedItems.removeElement(el.id);
    }

    document.sendCostumEvent("changed-selected");
    return false;
}
