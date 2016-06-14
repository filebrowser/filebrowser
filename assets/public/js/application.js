"use strict";

var selectedItems = [];

Array.prototype.removeElement = function(element) {
    var i = this.indexOf(element);
    if (i != -1) {
        this.splice(i, 1);
    }
}

document.addEventListener("DOMContentLoaded", function(event) {
    var items = document.getElementsByClassName('item');
    Array.from(items).forEach(link => {
        link.addEventListener('click', function(event) {
            var url = link.getElementsByTagName('a')[0].getAttribute('href');
            if (selectedItems.indexOf(url) == -1) {
                link.classList.add('selected');
                selectedItems.push(url);
            } else {
                link.classList.remove('selected');
                selectedItems.removeElement(url);
            }

            var event = new CustomEvent('changed-selected');
            document.dispatchEvent(event);
            return false;
        });
    });

    document.getElementById("open").addEventListener("click", openEvent);
    if (document.getElementById("back")) {
        document.getElementById("back").addEventListener("click", backEvent)
    };
    document.getElementById("delete").addEventListener("click", deleteEvent);
    document.getElementById("download").addEventListener("click", downloadEvent);
    return false;
});

var openEvent = function(event) {
    if (selectedItems.length) {


        return false;
    }

    window.open(window.location + "?raw=true");
    return false;
}

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

var deleteEvent = function(event) {
    Array.from(selectedItems).forEach(item => {
        var request = new XMLHttpRequest();
        request.open("DELETE", item);
        request.send();
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                if (request.status != 200) {
                    alert("something wrong happened!");
                    return false;
                }

                selectedItems.removeElement(item);
                alert(item + " deleted");
                // Add removing animation
            }
        }
    });
    return false;
}

var downloadEvent = function(event) {
    if (selectedItems.length) {
        Array.from(selectedItems).forEach(item => {
            window.open(item + "?download=true");
        });
        return false;
    }

    window.open(window.location + "?download=true");
    return false;
}

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

        return false;
    }

    toolbar.classList.remove("enabled");
    return false;
});
