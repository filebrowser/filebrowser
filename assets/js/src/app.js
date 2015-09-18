$(document).ready(function() {
  $(document).pjax('a', '#main');
});

$(document).on('ready pjax:success', function() {
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

  // Submites any form in the page in JSON format
  $('form').submit(function(event) {
    event.preventDefault();

    var data = JSON.stringify($(this).serializeJSON()),
      url = $(this).attr('action'),
      button = $(this).find("input[type=submit]:focus");

    $.ajax({
      type: 'POST',
      url: url,
      data: data,
      headers: {
        'X-Regenerate': button.data("regenerate")
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