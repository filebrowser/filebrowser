$(document).ready(function() {
  $(document).pjax('a', '#main');
});

$(document).on('ready pjax:success', function() {
  $('textarea.auto-size').textareaAutoSize();
  // Starts the perfect scroolbar plugin
  $('.scroll').perfectScrollbar();

  // Toggles between preview and editing mode
  $("#preview").click(function(event) {
    event.preventDefault();

    var preview = $("#preview-area"),
      editor = $('.editor textarea');

    if ($(this).data("previewing") == "true") {
      preview.hide();
      editor.fadeIn();
      $(this).data("previewing", "false");

      notification({
        text: "Think, relax and do the better you can!",
        type: 'information',
        timeout: 2000
      });
    } else {
      var converter = new showdown.Converter(),
        text = editor.val(),
        html = converter.makeHtml(text);

      editor.hide();
      preview.html(html).fadeIn();
      $(this).data("previewing", "true");

      notification({
        text: "This is how your post looks like.",
        type: 'information',
        timeout: 2000
      });
    }

    return false;
  });

  if ($('#content-area')[0]) {
    var myCodeMirror = CodeMirror.fromTextArea($('#content-area')[0], {
      mode: 'markdown',
      theme: 'mdn-like',
      lineWrapping: true,
      lineNumbers: false,
      scrollbarStyle: null
    });
  }

  // Submites any form in the page in JSON format
  $('form').submit(function(event) {
    event.preventDefault();

    var data = JSON.stringify($(this).serializeJSON()),
      button = $(this).find("input[type=submit]:focus");

    console.log(data)

    $.ajax({
      type: 'POST',
      url: window.location,
      data: data,
      headers: {
        'X-Regenerate': button.data("regenerate"),
        'X-Content-Type': button.data("type")
      },
      dataType: 'json',
      encode: true,
    }).done(function(data) {
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
  });

  // Log out the user sending bad credentials to the server
  $("#logout").click(function(e) {
    e.preventDefault();
    $.ajax({
      type: "GET",
      url: "/admin",
      async: false,
      username: "username",
      password: "password",
      headers: {
        "Authorization": "Basic xxx"
      }
    }).fail(function() {
      window.location = "/";
    });

    return false;
  });

  // Adds one more field to the current group
  // TODO: improve this function. add group/field/array/obj
  $(".add").click(function(e) {
    e.preventDefault();
    fieldset = $(this).closest("fieldset");
    fieldset.append("<input name=\"" + fieldset.attr("name") + "\" id=\"" + fieldset.attr("name") + "\" value=\"\"></input><br>");
    return false;
  });
});

$(document).on('pjax:send', function() {
  $('#loading').fadeIn()
})
$(document).on('pjax:complete', function() {
  $('#loading').fadeOut()
})