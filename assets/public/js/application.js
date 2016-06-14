"use strict";

var selectedItems = [];

document.addEventListener("DOMContentLoaded", function(event) {
    var items = document.getElementsByTagName('tr');
    Array.from(items).forEach(link => {
        link.addEventListener('click', function(event) {
            var url = link.getElementsByTagName('a')[0].getAttribute('href');
            if (selectedItems.indexOf(url) == -1) {
                link.classList.add('selected');
                selectedItems.push(url);
            } else {
                link.classList.remove('selected');
                var i = selectedItems.indexOf(url);
                if (i != -1) {
                    selectedItems.splice(i, 1);
                }
            }

            var event = new CustomEvent('changed-selected');
            document.dispatchEvent(event);
            return false;
        });
    });

    document.getElementById("back").addEventListener("click", backEvent);
    document.getElementById("delete").addEventListener("click", deleteEvent);
    document.getElementById("download").addEventListener("click", downloadEvent);
    return false;
});

var backEvent = function(event) {
  var items = document.getElementsByTagName('tr');
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

        alert(item + " deleted");
        // Add removing animation
      }
    }
  });
  return false;
}

var downloadEvent = function(event) {
  Array.from(selectedItems).forEach(item => {
    window.open(item + "?download=true");
  });
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
