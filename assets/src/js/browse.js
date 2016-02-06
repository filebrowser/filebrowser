$(document).on('page:browse', function() {
  $('body').off('click', '.rename').on('click', '.rename', function(event) {
    event.preventDefault();
    button = $(this);

    var filename = prompt("New file name:");

    if (filename == "") {
      return false;
    }

    if (filename.substring(0, 1) != "/") {
        filename = window.location.pathname.replace("/admin/browse/", "") + '/' + filename;
    }

    var content = '{"filename": "' + filename + '"}';

    $.ajax({
      type: 'PUT',
      url: button.data("file"),
      data: content,
      dataType: 'json',
      encode: true
    }).done(function(data) {
      $.pjax({
        url: window.location.pathname,
        container: '#content'
      });
      notification({
        text: button.data("message"),
        type: 'success',
        timeout: 5000
      });
    }).fail(function(data) {
      notification({
        text: 'Something went wrong.',
        type: 'error'
      });
      console.log(data);
    });

    return false;
  });

  $('body').off('click', '.delete').on('click', '.delete', function(event) {
    event.preventDefault();
    button = $(this);

    $.ajax({
      type: 'DELETE',
      url: button.data("file")
    }).done(function(data) {
      button.parent().parent().fadeOut();
      notification({
        text: button.data("message"),
        type: 'success',
        timeout: 5000
      });
    }).fail(function(data) {
      notification({
        text: 'Something went wrong.',
        type: 'error'
      });
      console.log(data);
    });

    return false;
  });

  $('.new').off('click').click(function(event) {
    event.preventDefault();

    if ($(this).data("opened")) {
      $('#new-file').fadeOut(200);
      $(this).data("opened", false);
    } else {
      $('#new-file').fadeIn(200);
      $(this).data("opened", true);
    }

    return false;
  });

  $('#new-file').on('keypress', 'input', function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      var value = $(this).val(),
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

      var content = '{"filename": "' + filename + '", "archetype": "' + archetype + '"}';

      $.ajax({
        type: 'POST',
        url: window.location.pathname,
        data: content,
        dataType: 'json',
        encode: true,
      }).done(function(data) {
        notification({
          text: "File created successfully.",
          type: 'success',
          timeout: 5000
        });

        $.pjax({
          url: window.location.pathname.replace("browse", "edit") + filename,
          container: '#content'
        })
      }).fail(function(data) {
        // error types
        notification({
          text: 'Something went wrong.',
          type: 'error'
        });
        console.log(data);
      });

      return false;
    }
  });

  $("#upload").click(function(event) {
    event.preventDefault();
    $('.actions input[type="file"]').click();
    return false;
  });

  $('input[type="file"]').on('change', function(event) {
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
});
