'use strict';

var templates = [];

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

function addFrontMatterItem(event) {
    event.preventDefault();
    let clone;
    let temp = document.getElementById(tempID)
    if (temp) {
        temp.remove();
    }

    let block = this.parentNode,
        type = block.dataset.type,
        id = block.id;

    // If the block is an array
    if (type === "array") {
        let fieldID = id + "[]",
            input = fieldID,
            count = block.querySelectorAll('.group > div').length;

        input = input.replace(/\[/, '\\[');
        input = input.replace(/\]/, '\\]');

        let fieldsets = block.getElementsByTagName("fieldset");

        if (fieldsets.length > 0) {
            let newtype = fieldsets[0].dataset.type,
                bid = id + "[" + fieldsets.length + "]",
                name = fieldsets.length;
            
            clone = document.importNode(templates.base.content, true);
            clone.querySelector('fieldset').id = bid;
            clone.querySelector('fieldset').dataset.type = newtype;
            clone.querySelector('h3').innerHTML = name;
            clone.querySelector('.delete').dataset.delete = bid;
            clone.querySelector('.delete').addEventListener('click', deleteFrontMatterItem);
            clone.querySelector('.add').addEventListener('click', addFrontMatterItem);
            block.appendChild(clone);
        } else {
            clone = document.importNode(templates.arrayItem.content, true);
            clone.querySelector('[data-type="array-item"]').id = `${fieldID}-${count}`;
            clone.querySelector('input').name = fieldID;
            clone.querySelector('input').id = fieldID;
            clone.querySelector('div.action').dataset.delete = `${fieldID}-${count}`;
            clone.querySelector('div.action').addEventListener('click', deleteFrontMatterItem);
            block.querySelector('.group').appendChild(clone)
            document.getElementById(`${fieldID}-${count}`).querySelector('input').focus();
        }
    }

    if (type == "object" || type == "parent") {    
        clone = document.importNode(templates.temporary.content, true);
        clone.querySelector('.group').id = tempID;
        clone.querySelector('.block').id = tempID;
        clone.querySelector('input').name = tempID;

        if (type == "parent") {
            document.querySelector('.frontmatter').insertBefore(clone, document.querySelector('div.button.add'));
        } else {
            block.insertBefore(clone, block.querySelector('.group'));
        }

        let temp = document.getElementById(tempID),
            input = temp.querySelector('input');
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

                let name = value.substring(0, value.lastIndexOf(':')),
                    newtype = value.substring(value.lastIndexOf(':') + 1, value.length);
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
                        clone = document.importNode(templates.base.content, true);
                        clone.querySelector('fieldset').id = bid;
                        clone.querySelector('fieldset').dataset.type = newtype;
                        clone.querySelector('h3').innerHTML = name;
                        clone.querySelector('.delete').dataset.delete = bid;
                        clone.querySelector('.delete').addEventListener('click', deleteFrontMatterItem);
                        clone.querySelector('.add').addEventListener('click', addFrontMatterItem);

                        if (type == "parent") {
                            document.querySelector('.frontmatter').insertBefore(clone, document.querySelector('div.button.add'));
                        } else {
                            block.appendChild(clone);
                        }
                        
                        break;
                    default:
                        let group = block.querySelector('.group');

                        if (group == null) {
                            block.insertAdjacentHTML('afterbegin', '<div class="group"></div>');
                            group = block.querySelector('.group');
                        }
                        
                        clone = document.importNode(templates.objectItem.content, true);
                        clone.querySelector('.block').id = `block-${bid}`;
                        clone.querySelector('.block').dataset.content = bid;
                        clone.querySelector('label').for = bid;
                        clone.querySelector('label').innerHTML = name;
                        clone.querySelector('input').name = bid;
                        clone.querySelector('input').id = bid;
                        clone.querySelector('.action').dataset.delete = `block-${bid}`;
                        clone.querySelector('.action').addEventListener('click', deleteFrontMatterItem);
                
                        group.appendChild(clone)
                        document.getElementById(bid).focus();
                }
            }
        });
    }

    return false;
}

document.addEventListener("DOMContentLoaded", (event) => {
    textareaAutoGrow();
    
    templates.array = document.getElementById("array-template");
    templates.arrayItem = document.getElementById("array-item-template");
    templates.base = document.getElementById('base-template');
    templates.objectItem = document.getElementById("object-item-template");
    templates.temporary = document.getElementById('temporary-template');

    let container = document.getElementById('editor'),
        button = document.querySelector('#submit span:first-child'),
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