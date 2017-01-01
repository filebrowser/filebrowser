'use strict';

var listing = {
    selectMultiple: false
};

listing.reload = function(callback) {
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

listing.itemDragStart = function(event) {
    let el = event.target;

    for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
            el = el.parentElement;
        }
    }

    event.dataTransfer.setData("id", el.id);
    event.dataTransfer.setData("name", el.querySelector('.name').innerHTML);
}

listing.itemDragOver = function(event) {
    event.preventDefault();
    let el = event.target;

    for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
            el = el.parentElement;
        }
    }

    el.style.opacity = 1;
}

listing.itemDrop = function(e) {
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
                listing.reload();
            }
        }
    }
}

listing.documentDrop = function(event) {
    event.preventDefault();
    let dt = event.dataTransfer,
        files = dt.files,
        el = event.target,
        items = document.getElementsByClassName('item');

    for (let i = 0; i < 5; i++) {
        if (el != null && !el.classList.contains('item')) {
            el = el.parentElement;
        }
    }

    if (files.length > 0) {
        if (el != null && el.classList.contains('item') && el.dataset.dir == "true") {
            listing.handleFiles(files, el.querySelector('.name').innerHTML + "/");
            return;
        }

        listing.handleFiles(files, "");
    } else {
        Array.from(items).forEach(file => {
            file.style.opacity = 1;
        });
    }
}

listing.rename = function(event) {
    if (!selectedItems.length || selectedItems.length > 1) {
        return false;
    }

    let item = document.getElementById(selectedItems[0]);

    if (item.classList.contains('disabled')) {
        return false;
    }

    let link = item.dataset.url,
        name = item.querySelector('.name').innerHTML;

    let submit = (event) => {
        event.preventDefault();

        let newName = event.currentTarget.querySelector('input').value,
            newLink = removeLastDirectoryPartOf(toWebDavURL(link)) + "/" + newName,
            html = buttons.rename.querySelector('i').changeToLoading(),
            request = new XMLHttpRequest();

        request.open('MOVE', toWebDavURL(link));
        request.setRequestHeader('Destination', newLink);
        request.send();
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                if (request.status != 201 && request.status != 204) {
                    span.innerHTML = name;
                } else {
                    closePrompt(event);

                    listing.reload(() => {
                        newName = btoa(newName);
                        selectedItems = [newName];
                        document.getElementById(newName).setAttribute("aria-selected", true);
                        listing.handleSelectionChange();
                    });
                }

                buttons.rename.querySelector('i').changeToDone((request.status != 201 && request.status != 204), html);
            }
        }

        return false;
    }

    let clone = document.importNode(templates.question.content, true);
    clone.querySelector('h3').innerHTML = 'Rename';
    clone.querySelector('input').value = name;
    clone.querySelector('.ok').innerHTML = 'Rename';
    clone.querySelector('form').addEventListener('submit', submit);

    document.querySelector('body').appendChild(clone)
    document.querySelector('.overlay').classList.add('active');
    document.querySelector('.prompt').classList.add('active');

    return false;
}

listing.handleFiles = function(files, base) {
    let button = document.getElementById("upload"),
        html = button.querySelector('i').changeToLoading();

    for (let i = 0; i < files.length; i++) {
        let request = new XMLHttpRequest();
        request.open('PUT', toWebDavURL(window.location.pathname + base + files[i].name));

        request.send(files[i]);
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                if (request.status == 201) {
                    listing.reload();
                }

                button.querySelector('i').changeToDone((request.status != 201), html);
            }
        }
    }

    return false;
}

listing.unselectAll = function() {
    let items = document.getElementsByClassName('item');
    Array.from(items).forEach(link => {
        link.setAttribute("aria-selected", false);
    });

    selectedItems = [];

    listing.handleSelectionChange();
    return false;
}

listing.handleSelectionChange = function(event) {
    listing.redefineDownloadURLs();

    let selectedNumber = selectedItems.length,
        fileAction = document.getElementById("file-only");

    if (selectedNumber) {
        fileAction.classList.remove("disabled");

        if (selectedNumber > 1) {
            buttons.open.classList.add("disabled");
            buttons.rename.classList.add("disabled");
        }

        if (selectedNumber == 1) {
            if (document.getElementById(selectedItems[0]).dataset.dir == "true") {
                buttons.open.classList.add("disabled");
            } else {
                buttons.open.classList.remove("disabled");
            }

            buttons.rename.classList.remove("disabled");
        }

        return false;
    }

    fileAction.classList.add("disabled");
    return false;
}

listing.redefineDownloadURLs = function() {
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

listing.openItem = function(event) {
    window.location = event.currentTarget.dataset.url;
}

listing.selectItem = function(event) {
    let el = event.currentTarget;

    if (selectedItems.length != 0) event.preventDefault();
    if (selectedItems.indexOf(el.id) == -1) {
        if (!event.ctrlKey && !listing.selectMultiple) listing.unselectAll();

        el.setAttribute("aria-selected", true);
        selectedItems.push(el.id);
    } else {
        el.setAttribute("aria-selected", false);
        selectedItems.removeElement(el.id);
    }

    listing.handleSelectionChange();
    return false;
}

listing.newFileButton = function(event) {
    event.preventDefault();

    let clone = document.importNode(templates.question.content, true);
    clone.querySelector('h3').innerHTML = 'New file';
    clone.querySelector('p').innerHTML = 'End with a trailing slash to create a dir.';
    clone.querySelector('.ok').innerHTML = 'Create';
    clone.querySelector('form').addEventListener('submit', listing.newFilePrompt);

    document.querySelector('body').appendChild(clone)
    document.querySelector('.overlay').classList.add('active');
    document.querySelector('.prompt').classList.add('active');
}

listing.newFilePrompt = function(event) {
    event.preventDefault();

    let button = document.getElementById('new'),
        html = button.querySelector('i').changeToLoading(),
        request = new XMLHttpRequest(),
        name = event.currentTarget.querySelector('input').value;

    request.open((name.endsWith("/") ? "MKCOL" : "PUT"), toWebDavURL(window.location.pathname + name));
    request.send();
    request.onreadystatechange = function() {
        if (request.readyState == 4) {
            button.querySelector('i').changeToDone((request.status != 201), html);
            listing.reload();
        }
    }

    closePrompt(event);
    return false;
}

listing.updateColumns = function(event) {
    let columns = Math.floor(document.getElementById('listing').offsetWidth / 300),
        items = getCSSRule('#listing.mosaic .item');

    items.style.width = `calc(${100/columns}% - 1em)`;
}

// Keydown events
window.addEventListener('keydown', (event) => {
    if (event.keyCode == 27) {
        listing.unselectAll();

        if (document.querySelectorAll('.prompt').length) {
            closePrompt(event);
        }
    }

    if (event.keyCode == 113) {
        listing.rename();
    }

    if (event.ctrlKey || event.metaKey) {
        console.log("hey")
        switch (String.fromCharCode(event.which).toLowerCase()) {
            case 's':
                event.preventDefault();
                window.location = "?download=true"
        }
    }
});

window.addEventListener("resize", () => {
    listing.updateColumns();
});

document.addEventListener('DOMContentLoaded', event => {
    listing.updateColumns();

    buttons.rename = document.getElementById("rename");
    buttons.upload = document.getElementById("upload");
    buttons.new = document.getElementById('new');
    buttons.download = document.getElementById('download');

    document.getElementById('multiple-selection-activate').addEventListener('click', event => {
        listing.selectMultiple = true;
        clickOverlay.click();
        
        document.getElementById('multiple-selection').classList.add('active');
        document.querySelector('body').style.paddingBottom = "4em";
    })

    document.getElementById('multiple-selection-cancel').addEventListener('click', event => {
        listing.selectMultiple = false;

        document.querySelector('body').style.paddingBottom = "0";
        document.getElementById('multiple-selection').classList.remove('active');
    })

    if (user.AllowEdit) {
        buttons.rename.addEventListener("click", listing.rename);
    }
    
    let items = document.getElementsByClassName('item');

    if (user.AllowNew) {
        buttons.upload.addEventListener("click", (event) => {
            document.getElementById("upload-input").click();
        });

        buttons.new.addEventListener('click', listing.newFileButton);

        // Drag and Drop
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

        document.addEventListener("drop", listing.documentDrop, false);
    }
    
    let touches = {
        id: '',
        count: 0
    };
    
    Array.from(items).forEach(file => {
        file.addEventListener('touchstart', event => {
            if (touches.id != file.id) {
                touches.id = file.id;
                touches.count = 1;
                
                return;
            }
            
            touches.count++;
            
            if (touches.count > 1) {
                window.location = file.dataset.url;
            }
        });
    });
});