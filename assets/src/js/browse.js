// When the page Browse is opened
$(document).on('page:browse', function() {
  var foreground = '#foreground';

  /* DELETE FILE */
  var removeForm = 'form#delete';
  var removeItem = null;

  $('#content').on('click', '.delete', function(event) {
    event.preventDefault();

    // Gets the information about the file the user wants to delete
    removeItem = new Object();
    removeItem.url = $(this).data("file");
    removeItem.row = $(this).parent().parent();
    removeItem.filename = $(removeItem.row).find('.filename').text();

    // Shows the remove form and the foreground
    $(removeForm).find('span').text(removeItem.filename);
    $(removeForm).fadeIn(200)
    $(foreground).fadeIn(200);

    return false;
  });

  $('#content').on('submit', removeForm, function(event) {
    event.preventDefault();

    // Checks if the item to remove is defined
    if (removeItem == null) {
      notification({
        text: "Something is wrong with your form.",
        type: "error"
      });
      return false;
    }

    // Makes the DELETE request to the server
    var request = new XMLHttpRequest();
    request.open("DELETE", removeItem.url);
    request.send();
    request.onreadystatechange = function() {
      if (request.readyState == 4) {
        var response = JSON.parse(request.responseText),
          type = "success",
          timeout = 5000;

        $(foreground).fadeOut(200);
        $(removeForm).fadeOut(200);
        $(removeItem.row).fadeOut(200);

        if (request.status != 200) {
          type = "error";
          timeout = false;
        }

        notification({
          text: response.message,
          type: type,
          timeout: timeout
        });

        removeItem = null;
      }
    }

    return false;
  });

  /* FILE UPLOAD */

  $('#content').on('change', 'input[type="file"]', function(event) {
    event.preventDefault();
    files = event.target.files;

    $('#loading').fadeIn();

    // Create a formdata object and add the files
    var data = new FormData();
    $.each(files, function(key, value) {
      data.append(key, value);
    });

    $.ajax({
      url: window.location.pathname,
      type: 'POST',
      data: data,
      cache: false,
      dataType: 'json',
      headers: {
        'X-Upload': 'true',
      },
      processData: false,
      contentType: false,
    }).done(function(data) {
      notification({
        text: "File(s) uploaded successfully.",
        type: 'success',
        timeout: 5000
      });

      $('#loading').fadeOut();

      $.pjax({
        url: window.location.pathname,
        container: '#content'
      })
    }).fail(function(data) {
      $('#loading').fadeOut();

      notification({
        text: 'Something went wrong.',
        type: 'error'
      });
      console.log(data);
    });


    return false;
  });

  $('#content').on('click', '#upload', function(event) {
    event.preventDefault();
    $('.actions input[type="file"]').click();
    return false;
  });

  /* NEW FILE */
  var createForm = 'form#new',
    createInput = createForm + ' input[type="text"]';

  $('#content').on('click', '.new', function(event) {
    event.preventDefault();

    $(foreground).fadeIn(200);
    $(createForm).fadeIn(200);

    return false;
  });

  $('#content').on('keypress', createInput, function(event) {
    // If it's "enter" key, submit the
    if (event.keyCode == 13) {
      event.preventDefault();
      $(createForm).submit();
      return false;
    }
  });

  $('#content').on('submit', createForm, function(event) {
    event.preventDefault();

    var value = $(createInput).val(),
      splited = value.split(":"),
      filename = "",
      archetype = "";

    if (value == "") {
      notification({
        text: "You have to write something. If you want to close the box, click the button again.",
        type: 'warning',
        timeout: 5000
      });

      return false;
    } else if (splited.length == 1) {
      filename = value;
    } else if (splited.length == 2) {
      filename = splited[0];
      archetype = splited[1];
    } else {
      notification({
        text: "Hmm... I don't understand you. Try writing something like 'name[:archetype]'.",
        type: 'error'
      });

      return false;
    }

    var content = {
      filename: filename,
      archetype: archetype
    }

    var request = new XMLHttpRequest();
    request.open("POST", window.location.pathname);
    request.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    request.send(JSON.stringify(content));
    request.onreadystatechange = function() {
      if (request.readyState == 4) {
        var response = JSON.parse(request.responseText);
        var type = "success";
        var timeout = 5000;

        if (request.status != 200) {
          type = "error";
          timeout = false;
        }

        notification({
          text: response.message,
          type: type,
          timeout: timeout
        });

        if (request.status == 200) {
          $.pjax({
            url: response.location,
            container: '#content'
          })
        }
      }
    }

    return false;
  });

  /* RENAME FILE */
  var renameForm = 'form#rename',
    renameInput = renameForm + ' input[type="text"]',
    renameItem = null;

  $('#content').on('click', '.rename', function(event) {
    event.preventDefault();
    renameItem = $(this).parent().parent().find('.filename').text();
    $(foreground).fadeIn(200);
    $(renameForm).fadeIn(200);
    $(renameForm).find('span').text(renameItem);
    $(renameForm).find('input[type="text"]').val(renameItem);
    return false;
  });

  $('#content').on('keypress', renameInput, function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      $(renameForm).submit();
      return false;
    }
  });

  $('#content').on('submit', renameForm, function(event) {
    event.preventDefault();

    var filename = $(this).find('input[type="text"]').val();
    if (filename === "") {
      return false;
    }

    if (filename.substring(0, 1) != "/") {
      filename = window.location.pathname.replace(basePath + "/browse/", "") + '/' + filename;
    }

    var content = {
      filename: filename
    };

    var request = new XMLHttpRequest();
    request.open("PUT", renameItem);
    request.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    request.send(JSON.stringify(content));
    request.onreadystatechange = function() {
      if (request.readyState == 4) {
        var response = JSON.parse(request.responseText),
          type = "success",
          timeout = 5000;

        if (request.status != 200) {
          type = "error";
          timeout = false;
        }

        $.pjax({
          url: window.location.pathname,
          container: '#content'
        });

        notification({
          text: response.message,
          type: type,
          timeout: timeout
        });

        renameItem = null;
      }
    }

    return false;
  });

  /* GIT ACTIONS */
  var gitButton = 'button.git',
    gitForm = 'form#git',
    gitInput = gitForm + ' input[type="text"]';

  $('#content').on('click', gitButton, function(event) {
    event.preventDefault();
    $(foreground).fadeIn(200);
    $(gitForm).fadeIn(200);
    return false;
  });

  $('#content').on('keypress', gitInput, function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      $(gitForm).submit();
      return false;
    }
  });

  $('#content').on('submit', gitForm, function(event) {
    event.preventDefault();

    var value = $(this).find('input[type="text"]').val();

    if (value == "") {
      notification({
        text: "You have to write something. If you want to close the box, click outside of it.",
        type: 'warning',
        timeout: 5000
      });

      return false;
    }

    var request = new XMLHttpRequest();
    request.open("POST", basePath + "/git");
    request.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    request.send(JSON.stringify({
      command: value
    }));
    request.onreadystatechange = function() {
      if (request.readyState == 4) {
        var data = JSON.parse(request.responseText);

        if (request.status == 200) {
          notification({
            text: data.message,
            type: "success",
            timeout: 5000
          });

          $(gitForm).fadeOut(200);
          $(foreground).fadeOut(200);

          $.pjax({
            url: window.location.pathname,
            container: '#content'
          });
        } else {
          notification({
            text: data.message,
            type: "error"
          });
        }
      }
    }

    return false;
  });

  /* $(foreground) AND STUFF */

  $('#content').on('click', '.close', function(event) {
    event.preventDefault();
    $(this).parent().parent().fadeOut(200);
    $(foreground).click();
    return false;
  });

  $('#content').on('click', foreground, function(event) {
    event.preventDefault();
    $(foreground).fadeOut(200);
    $(createForm).fadeOut(200);
    $(renameForm).fadeOut(200);
    $(removeForm).fadeOut(200);
    $(gitForm).fadeOut(200);
    return false;
  });
});
