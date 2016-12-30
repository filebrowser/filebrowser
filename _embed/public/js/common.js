'use strict';

var tempID = "_fm_internal_temporary_id",
    buttons = {};

// Removes an element, if exists, from an array
Array.prototype.removeElement = function(element) {
    var i = this.indexOf(element);
    if (i != -1) this.splice(i, 1);
}

// Replaces an element inside an array by another
Array.prototype.replaceElement = function(oldElement, newElement) {
    var i = this.indexOf(oldElement);
    if (i != -1) this[i] = newElement;
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
    let element = this,
        originalText = element.innerHTML;

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

        if (selectedItems.length == 0 && document.getElementById('listing')) document.sendCostumEvent('changed-selected');
    }

    let secondStep = () => {
        this.style.opacity = 0;
        setTimeout(thirdStep, 200);
    }

    let firstStep = () => {
        this.innerHTML = '<i class="material-icons">done</i>';
        if (error) {
          this.innerHTML = '<i class="material-icons">close</i>';
        }

        this.style.opacity = 1;

        setTimeout(secondStep, 1000);
    }

    setTimeout(firstStep, 200);
    return false;
}

function getCSSRule(ruleName) {
    ruleName = ruleName.toLowerCase();
    var result = null,
        find = Array.prototype.find;

    find.call(document.styleSheets, styleSheet => {
        result = find.call(styleSheet.cssRules, cssRule => {
            return cssRule instanceof CSSStyleRule &&
                cssRule.selectorText.toLowerCase() == ruleName;
        });
        return result != null;
    });
    return result;
}

function toWebDavURL(url) {
    return window.location.origin + url.replace(baseURL + "/", webdavURL + "/");
}

// Remove the last directory of an url
var removeLastDirectoryPartOf = function(url) {
    var arr = url.split('/');
    arr.pop();
    return (arr.join('/'));
}

/* * * * * * * * * * * * * * * *
 *                             *
 *            EVENTS           *
 *                             *
 * * * * * * * * * * * * * * * */

// Prevent Default event
var preventDefault = function(event) {
    event.preventDefault();
}

function logoutEvent(event) {
    let request = new XMLHttpRequest();
    request.open('GET', window.location.pathname, true, "username", "password");
    request.send();
    request.onreadystatechange = function() {
        if (request.readyState == 4) {
            window.location = "/";
        }
    }
}

function openEvent(event) {
    if (event.currentTarget.classList.contains('disabled')) {
        return false;
    }

    let link = '?raw=true';

    if (selectedItems.length) {
        link = document.getElementById(selectedItems[0]).dataset.url + link;
    } else {
        link = window.location + link;
    }

    window.open(link);
    return false;
}

// Handles the delete button event
function deleteEvent(event) {
    let single = false;

    if (!selectedItems.length) {
        selectedItems = [window.location.pathname];
        single = true;
    }

    Array.from(selectedItems).forEach(id => {
        let request = new XMLHttpRequest(),
            html = buttons.delete.changeToLoading(),
            el = document.getElementById(id),
            url = el.dataset.url;

        request.open('DELETE', toWebDavURL(url));
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                if (request.status == 204) {
                    if (single) {
                        window.location.pathname = removeLastDirectoryPartOf(window.location.pathname);
                    } else {
                        el.remove();
                        selectedItems.removeElement(id);
                    }
                }

                buttons.delete.changeToDone(request.status != 204, html);
            }
        }
        r.send();
    });

    return false;
}

var searchEvent = function(event) {
    let value = this.value,
        search = document.getElementById('search'),
        scrollable = document.querySelector('#search > div'),
        box = document.querySelector('#search > div div');

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

/* * * * * * * * * * * * * * * *
 *                             *
 *           BOOTSTRAP         *
 *                             *
 * * * * * * * * * * * * * * * */

document.addEventListener("DOMContentLoaded", function(event) {
    buttons.logout = document.getElementById("logout");
    buttons.open = document.getElementById("open");
    buttons.delete = document.getElementById("delete");

    // Attach event listeners
    buttons.logout.addEventListener("click", logoutEvent);
    buttons.open.addEventListener("click", openEvent);

    if (user.AllowEdit) {
        buttons.delete.addEventListener("click", deleteEvent);
    }

    return false;
});
