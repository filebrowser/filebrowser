$(document).on('page:browse', function() {
  var foreground = '#foreground';

  /* DELETE FILE */

  var remove = new Object();
  remove.selector = 'form#delete';
  remove.form = $(remove.selector);
  remove.row = '';
  remove.button = '';
  remove.url = '';

  $('#content').on('click', '.delete', function(event) {
    event.preventDefault();
    remove.button = $(this);
    remove.row = $(this).parent().parent();
    $(foreground).fadeIn(200);
    remove.url = remove.row.find('.filename').text();
    remove.form.find('span').text(remove.url);
    remove.form.fadeIn(200);
    return false;
  });

  $('#content').on('submit', remove.selector, function(event) {
    event.preventDefault();

    var request = new XMLHttpRequest();
    request.open("DELETE", remove.button.data("file"));
    request.send();
    request.onreadystatechange = function() {
      if (request.readyState == 4) {
        var response = JSON.parse(request.responseText),
          type = "success",
          timeout = 5000;

        $(foreground).fadeOut(200);
        remove.form.fadeOut(200);
        remove.row.fadeOut(200);

        if (request.status != 200) {
          type = "error";
          timeout = false;
        }

        notification({
          text: response.message,
          type: type,
          timeout: timeout
        });
      }
    }

    return false;
  });

  /* FILE UPLOAD */

  $('#content').on('change', 'input[type="file"]', function(event) {
    event.preventDefault();
    files = event.target.files;

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

      $.pjax({
        url: window.location.pathname,
        container: '#content'
      })
    }).fail(function(data) {
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

  var create = new Object();
  create.selector = 'form#new';
  create.form = $(create.selector);
  create.input = create.selector + ' input[type="text"]';
  create.button = '';
  create.url = '';

  $('#content').on('click', '.new', function(event) {
    event.preventDefault();
    create.button = $(this);
    $(foreground).fadeIn(200);
    create.form.fadeIn(200);
    return false;
  });

  $('#content').on('keypress', create.input, function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      $(create.form).submit();
      return false;
    }
  });

  $('#content').on('submit', create.selector, function(event) {
    event.preventDefault();

    var value = create.form.find('input[type="text"]').val(),
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

  var rename = new Object();
  rename.selector = 'form#rename';
  rename.form = $(rename.selector);
  rename.input = rename.selector + ' input[type="text"]';
  rename.button = '';
  rename.url = '';

  $('#content').on('click', '.rename', function(event) {
    event.preventDefault();
    rename.button = $(this);

    $(foreground).fadeIn(200);
    rename.url = $(this).parent().parent().find('.filename').text();
    rename.form.fadeIn(200);
    rename.form.find('span').text(rename.url);
    rename.form.find('input[type="text"]').val(rename.url);

    return false;
  });

  $('#content').on('keypress', rename.input, function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      $(rename.form).submit();
      return false;
    }
  });

  $('#content').on('submit', rename.selector, function(event) {
    event.preventDefault();

    var filename = rename.form.find('input[type="text"]').val();
    if (filename === "") {
      return false;
    }

    if (filename.substring(0, 1) != "/") {
      filename = window.location.pathname.replace("/admin/browse/", "") + '/' + filename;
    }

    var content = {
      filename: filename
    };

    var request = new XMLHttpRequest();
    request.open("PUT", rename.url);
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
      }
    }

    return false;
  });

  /* GIT ACTIONS */

  var git = new Object();
  git.selector = 'form#git';
  git.form = $(git.selector);
  git.input = git.selector + ' input[type="text"]';

  $('#content').on('click', 'button.git', function(event) {
    event.preventDefault();
    $(foreground).fadeIn(200);
    git.form.fadeIn(200);
    return false;
  });

  $('#content').on('keypress', git.input, function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      $(git.form).submit();
      return false;
    }
  });

  $('#content').on('submit', git.selector, function(event) {
    event.preventDefault();

    var value = git.form.find('input[type="text"]').val();

    if (value == "") {
      notification({
        text: "You have to write something. If you want to close the box, click outside of the box.",
        type: 'warning',
        timeout: 5000
      });

      return false;
    }

    var request = new XMLHttpRequest();
    request.open("POST", "/admin/git");
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
            type: "success"
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
    create.form.fadeOut(200);
    rename.form.fadeOut(200);
    remove.form.fadeOut(200);
    git.form.fadeOut(200);
    return false;
  });
});
