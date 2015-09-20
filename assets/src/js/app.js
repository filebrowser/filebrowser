$(document).ready(function() {
  $(document).pjax('a[data-pjax]', '#content');
});

$(document).on('ready pjax:success', function() {
  // Starts the perfect scroolbar plugin
  $('.scroll').perfectScrollbar();

  // Log out the user sending bad credentials to the server
  $("#logout").click(function(event) {
    event.preventDefault();
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

  // If it's editor page
  if ($(".editor")[0]) {
    editor = false;
    preview = $("#preview-area");
    textarea = $("#content-area");

    // If it has a textarea
    if (textarea[0]) {
      options = {
        mode: textarea.data("mode"),
        theme: 'mdn-like',
        lineWrapping: true,
        lineNumbers: true,
        scrollbarStyle: null
      }

      if (textarea.data("mode") == "markdown") {
        options.lineNumbers = false
      }

      editor = CodeMirror.fromTextArea(textarea[0], options);
      codemirror = $('.CodeMirror');

      // Toggles between preview and editing mode
      $("#preview").click(function(event) {
        event.preventDefault();

        // If it currently in the preview mode, hide the preview
        // and show the editor
        if ($(this).data("previewing") == "true") {
          preview.hide();
          codemirror.fadeIn();
          $(this).data("previewing", "false");
          notification({
            text: "Think, relax and do the better you can!",
            type: 'information',
            timeout: 2000
          });
        } else {
          // Copy the editor content to texteare
          editor.save()

          // If it's in editing mode, convert the markdown to html
          // and show it
          var converter = new showdown.Converter(),
            text = textarea.val(),
            html = converter.makeHtml(text);

          // Hide the editor and show the preview
          codemirror.hide();
          preview.html(html).fadeIn();
          $('pre code').each(function(i, block) {
            hljs.highlightBlock(block);
            return true;
          });

          $(this).data("previewing", "true");
          notification({
            text: "This is how your post looks like.",
            type: 'information',
            timeout: 2000
          });
        }

        return false;
      });
    }

    // Submites any form in the page in JSON format
    $('form').submit(function(event) {
      event.preventDefault();
      var data = JSON.stringify($(this).serializeJSON()),
        button = $(this).find("input[type=submit]:focus");

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

    // Adds one more field to the current group
    $(".add").click(function(event) {
      event.preventDefault();

      if ($("#new").length) {
        return false;
      }

      title = $(this).parent().parent();
      fieldset = title.parent();
      type = fieldset.data("type");
      name = fieldset.data("name");

      if (title.is('h1')) {
        fieldset = $('.sidebar .content');
        fieldset.prepend('<div id="ghost"></div>');
        title = $('#ghost');
        type = "object";
      }

      if (type == "object") {
        title.after('<input id="new" placeholder="Write the field name and press enter..."></input>');
        element = $("#new");

        $(element).keypress(function(event) {
          if (event.which == 13) {
            event.preventDefault();
            value = element.val();
            element.remove();

            if (value == "") {
              return false;
            }

            if (name == "undefined") {
              name = value
            } else {
              name = name + '[' + value + ']';
            }

            title.after('<input name="' + name + ':auto" id="' + name + '"></input><br>');
            title.after('<label for="' + name + '">' + value + ' <span class="actions"><button class="delete"><i class="fa fa-minus"></i></button></span></label>');

            return false;
          }
        });
      }

      if (type == "array") {
        name = name + "[]";
        title.after('<input name="' + name + ':auto" id="' + name + '"></input><br>');
      }

      return false;
    });

    $(".delete").click(function(event) {
      event.preventDefault();
      name = $(this).parent().parent().attr("for") || $(this).parent().parent().parent().attr("id");
      console.log(name)

      $('#' + name).fadeOut().remove();
      $('label[for="' + name + '"]').fadeOut().remove();
    });

    $('body').on('keypress', 'input', function(event) {
      if (event.keyCode == 13) {
        event.preventDefault();
        $('input[value="Save"]').focus().click();
        return false;
      }
    });
  }
});