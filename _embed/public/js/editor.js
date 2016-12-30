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

function addFrontMatterItem(event) {
    event.preventDefault();

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

document.addEventListener("DOMContentLoaded", (event) => {
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