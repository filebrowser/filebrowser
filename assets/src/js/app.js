$(document).ready(function() {
  $(document).pjax('a[data-pjax]', '#content');
});

$(document).on('ready pjax:success', function() {
  // Auto Grow Textarea
  function autoGrow(element) {
    this.style.height = "5px";
    this.style.height = (this.scrollHeight) + "px";
  }

  // Auto Grow textareas after loading
  $("textarea").each(autoGrow);
  // Auto Grow textareas when changing its content
  $('textarea').keyup(autoGrow);
  // Auto Grow textareas when resizing the window
  $(window).resize(function() {
    $("textarea").each(autoGrow);
  });

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

  if ($('main').hasClass('browse')) {
    $('.new').click(function(event) {
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
  }

  // If it's editor page
  if ($(".editor")[0]) {
    var editor = ace.edit("source-area");
    editor.getSession().setMode("ace/mode/markdown");
    editor.setOptions({
      wrap: true,
      maxLines: Infinity,
      theme: "ace/theme/github",
      showPrintMargin: false,
      fontSize: "1em"
    });


    preview = $("#preview-area");
    textarea = $("#content-area");

    $('body').on('keypress', 'input', function(event) {
      if (event.keyCode == 13) {
        event.preventDefault();
        $('input[value="Save"]').focus().click();
        return false;
      }
    });

    // Submites any form in the page in JSON format
    $('form').submit(function(event) {
      event.preventDefault();

      // Reset preview area and button to make sure it will
      // not serialize any form inside the preview
      $('#preview-area').html('').fadeOut();
      $('#preview').data("previewing", "false");
      $('.CodeMirror').fadeIn();

      // Save editor values
      if (typeof editor !== 'undefined' && editor) {
        editor.save();
      }

      var data = JSON.stringify($(this).serializeJSON()),
        button = $(this).find("input[type=submit]:focus");

      $.ajax({
        type: 'POST',
        url: window.location,
        data: data,
        headers: {
          'X-Regenerate': button.data("regenerate"),
          'X-Schedule': button.data("schedule"),
          'X-Content-Type': button.data("type")
        },
        dataType: 'json',
        encode: true,
        contentType: "application/json; charset=utf-8",
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
      defaultId = "new-admin-item-123";

      if ($("#" + defaultId).length) {
        return false;
      }

      fieldset = $(this).parent().parent();
      type = fieldset.data("type");
      name = fieldset.attr("id");
      newItem = "";

      // Main add button, below the title
      if (fieldset.is('div') && fieldset.hasClass("frontmatter")) {
        fieldset = $('.blocks');
        fieldset.append('<div class="block" id="' + defaultId + '"></div>');
        newItem = $("#" + defaultId);
        type = "object";
      }

      if (type == "array") {
        name = name + "[]";
        input = name;
        input = input.replace(/\[/, '\\[');
        input = input.replace(/\]/, '\\]');
        input = '#' + input;

        fieldset.append('<div id="' + name + '-' + $(input).length + '" data-type="array-item"><input name="' + name + ':auto" id="' + name + '"></input><span class="actions"> <button class="delete">&#8722;</button></span></div></div>');
      }

      if (type == "object") {
        newItem.html('<input id="name-' + defaultId + '" placeholder="Write the field name and press enter..."></input>');
        element = $("#name-" + defaultId);

        if (!document.cookie.replace(/(?:(?:^|.*;\s*)placeholdertip\s*\=\s*([^;]*).*$)|^.*$/, "$1")) {
          var date = new Date();
          date.setDate(date.getDate() + 365);
          document.cookie = 'placeholdertip=true; expires=' + date.toUTCString + '; path=/';

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
              newItem.remove();
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
              newItem.append('<input name="' + name + ':auto" id="' + name + '"></input><br>');
              newItem.prepend('<label for="' + name + '">' + value + '</label> <span class="actions"><button class="delete">&#8722;</button></span>');
            } else {
              var fieldset = "<fieldset id=\"{{ $value.Name }}\" data-type=\"{{ $value.Type }}\">\r\n<h3>{{ $value.Title }}\r\n<span class=\"actions\">\r\n<button class=\"add\"><i class=\"fa fa-plus\"><\/i><\/button>\r\n<button class=\"delete\"><i class=\"fa fa-minus\"><\/i><\/button>\r\n<\/span>\r\n<\/h3>\r\n<\/fieldset>";

              if (elements[1] == "array") {
                fieldset = fieldset.replace("{{ $value.Type }}", "array");
              } else {
                fieldset = fieldset.replace("{{ $value.Type }}", "object");
              }

              fieldset = fieldset.replace("{{ $value.Title }}", elements[0]);
              fieldset = fieldset.replace("{{ $value.Name }}", name);
              newItem.after(fieldset);
            }

            return false;
          }
        });
      }

      /*

      /* WHAT IS THIS FOR?
      if (title.is('h2')) {
        type = "object"
      }



      */

      return false;
    });

    // If it has a textarea
    if (textarea[0]) {
      options = {
        mode: textarea.data("mode"),
        theme: 'ttcn',
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
  }
});
