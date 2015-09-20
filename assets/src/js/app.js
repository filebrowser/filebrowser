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

  // Delete a file or a field in editor
  $("body").on('click', '.delete', function(event) {
    event.preventDefault();
    button = $(this);

    if (button.data("file") && confirm("Are you sure you want to delete this?")) {
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
    } else {
      name = button.parent().parent().attr("for") || button.parent().parent().attr("id") || button.parent().parent().parent().attr("id");
      name = name.replace(/\[/, '\\[');
      name = name.replace(/\]/, '\\]');
      console.log(name)

      $('label[for="' + name + '"]').fadeOut().remove();
      $('#' + name).fadeOut().remove();
    }

    return false;
  });

  // If it's editor page
  if ($(".editor")[0]) {
    editor = false;
    preview = $("#preview-area");
    textarea = $("#content-area");

    $('body').on('keypress', 'input', function(event) {
      if (event.keyCode == 13) {
        event.preventDefault();
        $('input[value="Save"]').focus().click();
        return false;
      }
    });

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

      return false;
    });

    // Adds one more field to the current group
    $("body").on('click', '.add', function(event) {
      event.preventDefault();

      if ($("#new").length) {
        return false;
      }

      title = $(this).parent().parent();
      fieldset = title.parent();
      type = fieldset.data("type");
      name = fieldset.attr("id");

      if (title.is('h1')) {
        fieldset = $('.frontmatter .container');
        fieldset.prepend('<div id="ghost"></div>');
        title = $('#ghost');
        type = "object";
      }

      if (type == "object") {
        title.after('<input id="new" placeholder="Write the field name and press enter..."></input>');
        element = $("#new");

        if (!Cookies.get('placeholdertip')) {
          Cookies.set('placeholdertip', 'true', {
            expires: 365
          });

          notification({
            text: 'Write the field name and then press enter. If you want to create an array or an object, end the name with ":array" or ":object".',
            type: 'information'
          });
        }

        $(element).keypress(function(event) {
          if (event.which == 13) {
            event.preventDefault();
            value = element.val();

            if (value == "") {
              element.remove();
              return false;
            }

            elements = value.split(":")

            if (elements.length > 2) {
              notification({
                text: "Invalid syntax. It must be 'name[:type]'.",
                type: 'error'
              });
              return false;
            }

            element.remove();

            if (name == "undefined") {
              name = elements[0]
            } else {
              name = name + '[' + elements[0] + ']';
            }

            if (elements.length == 1) {
              title.after('<input name="' + name + ':auto" id="' + name + '"></input><br>');
              title.after('<label for="' + name + '">' + value + ' <span class="actions"><button class="delete"><i class="fa fa-minus"></i></button></span></label>');
            } else {
              var fieldset = "<fieldset id=\"{{ $value.Name }}\" data-type=\"{{ $value.Type }}\">\r\n<h3>{{ $value.Title }}\r\n<span class=\"actions\">\r\n<button class=\"add\"><i class=\"fa fa-plus\"><\/i><\/button>\r\n<button class=\"delete\"><i class=\"fa fa-minus\"><\/i><\/button>\r\n<\/span>\r\n<\/h3>\r\n<\/fieldset>";

              if (elements[1] == "array") {
                fieldset = fieldset.replace("{{ $value.Type }}", "array");
              } else {
                fieldset = fieldset.replace("{{ $value.Type }}", "object");
              }

              fieldset = fieldset.replace("{{ $value.Title }}", elements[0]);
              fieldset = fieldset.replace("{{ $value.Name }}", name);
              title.after(fieldset);
            }

            return false;
          }
        });
      }

      if (type == "array") {
        name = name + "[]";
        input = name;
        input = input.replace(/\[/, '\\[');
        input = input.replace(/\]/, '\\]');
        input = '#' + input;

        title.after('<div id="' + name + '-' + $(input).length + '" data-type="array-item"><input name="' + name + ':auto" id="' + name + '"></input></div>');
      }

      return false;
    });
  }
});