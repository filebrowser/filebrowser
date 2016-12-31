'use strict';

function textareaAutoGrow() {
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

function deleteFrontMatterItem(event) {
    event.preventDefault();
    document.getElementById(this.dataset.delete).remove();
}

function makeFromBaseTemplate(id, type, name, parent) {
    let clone = document.importNode(templates.base.content, true);
    clone.querySelector('fieldset').id = id;
    clone.querySelector('fieldset').dataset.type = type;
    clone.querySelector('h3').innerHTML = name;
    clone.querySelector('.delete').dataset.delete = id;
    clone.querySelector('.delete').addEventListener('click', deleteFrontMatterItem);
    clone.querySelector('.add').addEventListener('click', addFrontMatterItem);

    if (parent.classList.contains("frontmatter")) {
        parent.insertBefore(clone, document.querySelector('div.button.add'));
        return
    }

    parent.appendChild(clone);
}

function makeFromArrayItemTemplate(id, number, parent) {
    let clone = document.importNode(templates.arrayItem.content, true);
    clone.querySelector('[data-type="array-item"]').id = `${id}-${number}`;
    clone.querySelector('input').name = id;
    clone.querySelector('input').id = id;
    clone.querySelector('div.action').dataset.delete = `${id}-${number}`;
    clone.querySelector('div.action').addEventListener('click', deleteFrontMatterItem);
    parent.querySelector('.group').appendChild(clone)
    document.getElementById(`${id}-${number}`).querySelector('input').focus();
}

function makeFromObjectItemTemplate(id, name, parent) {
    let clone = document.importNode(templates.objectItem.content, true);
    clone.querySelector('.block').id = `block-${id}`;
    clone.querySelector('.block').dataset.content = id;
    clone.querySelector('label').for = id;
    clone.querySelector('label').innerHTML = name;
    clone.querySelector('input').name = id;
    clone.querySelector('input').id = id;
    clone.querySelector('.action').dataset.delete = `block-${id}`;
    clone.querySelector('.action').addEventListener('click', deleteFrontMatterItem);

    parent.appendChild(clone)
    document.getElementById(id).focus();
}

function addFrontMatterItemPrompt(parent) {
    return function(event) {
        event.preventDefault();

        let value = event.currentTarget.querySelector('input').value;
        if (value === '') {
            return true;
        }

        closePrompt(event);

        let name = value.substring(0, value.lastIndexOf(':')),
            type = value.substring(value.lastIndexOf(':') + 1, value.length);

        if (type !== "" && type !== "array" && type !== "object") {
            name = value;
        }

        name = name.replace(' ', '_');

        let id = name;

        if (parent.id != '') {
            id = parent.id + "." + id;
        }

        if (type == "array" || type == "object") {
            if (parent.dataset.type == "parent") {
                makeFromBaseTemplate(bid, newtype, name, document.querySelector('.frontmatter'));
                return;
            }

            makeFromBaseTemplate(bid, newtype, name, block);
            return;
        }

        let group = parent.querySelector('.group');

        if (group == null) {
            parent.insertAdjacentHTML('afterbegin', '<div class="group"></div>');
            group = parent.querySelector('.group');
        }

        makeFromObjectItemTemplate(id, name, group);
    }
}

function addFrontMatterItem(event) {
    event.preventDefault();

    let parent = event.currentTarget.parentNode,
        type = parent.dataset.type;

    // If the block is an array
    if (type === "array") {
        let id = parent.id + "[]",
            count = parent.querySelectorAll('.group > div').length,
            fieldsets = parent.getElementsByTagName("fieldset");

        if (fieldsets.length > 0) {
            let itemType = fieldsets[0].dataset.type,
                itemID = parent.id + "[" + fieldsets.length + "]",
                itemName = fieldsets.length;

            makeFromBaseTemplate(itemID, itemType, itemName, parent);
        } else {
            makeFromArrayItemTemplate(id, count, parent);
        }

        return;
    }

    if (type == "object" || type == "parent") {
        let clone = document.importNode(templates.question.content, true);
        clone.querySelector('form').id = tempID;
        clone.querySelector('h3').innerHTML = 'New field';
        clone.querySelector('p').innerHTML = 'Write the field name and then press enter. If you want to create an array or an object, end the name with <code>:array</code> or <code>:object.</code>';
        clone.querySelector('.ok').innerHTML = 'Create';
        clone.querySelector('form').addEventListener('submit', addFrontMatterItemPrompt(parent));
        clone.querySelector('form').classList.add('active')
        document.querySelector('body').appendChild(clone);

        document.querySelector('.overlay').classList.add('active');
        document.getElementById(tempID).classList.add('active');
    }

    return false;
}

document.addEventListener("DOMContentLoaded", (event) => {
    textareaAutoGrow();

    templates.arrayItem = document.getElementById("array-item-template");
    templates.base = document.getElementById('base-template');
    templates.objectItem = document.getElementById("object-item-template");
    templates.temporary = document.getElementById('temporary-template');

    let container = document.getElementById('editor'),
        button = document.querySelector('#save'),
        kind = container.dataset.kind;

    if (kind != 'frontmatter-only') {
        let editor = document.getElementById('editor-source'),
            mode = editor.dataset.mode,
            textarea = document.querySelector('textarea[name="content"]'),
            aceEditor = ace.edit('editor-source');
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

        let html = button.changeToLoading(),
            request = new XMLHttpRequest();

        request.open("PUT", toWebDavURL(window.location.pathname));
        request.setRequestHeader('Kind', kind);
        request.send(JSON.stringify(data));
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                button.changeToDone((request.status != 201), html);
            }
        }
    }

    document.querySelector('#save').addEventListener('click', event => {
        event.preventDefault();
        saveContent();
    });

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