$(document).ready(function() {
  // Start pjax
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
    var mode = $("#source-area").data('mode');
    var editor = ace.edit("source-area");
    editor.getSession().setMode("ace/mode/" + mode);
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

    //TODO: reform this
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
      defaultID = "lorem-ipsum-sin-dolor-amet";

      if ($("#" + defaultID).length) {
        return false;
      }

      block = $(this).parent().parent();
      blockType = block.data("type");
      blockID = block.attr("id");

      // Main add button, after all blocks
      if (block.is('div') && block.hasClass("frontmatter")) {
        block = $('.blocks');
        block.append('<div class="block" id="' + defaultID + '"></div>');
        blockType = "object";
      }

      // If the Block Type is an array
      if (blockType == "array") {
        newID = blockID + "[]";
        input = blockID;
        input = input.replace(/\[/, '\\[');
        input = input.replace(/\]/, '\\]');
        block.append('<div id="' + newID + '-' + $('#' + input + ' > div').length + '" data-type="array-item"><input name="' + newID + ':auto" id="' + newID + '"></input><span class="actions"> <button class="delete">&#8722;</button></span></div></div>');
      }

      // If the Block is an object
      if (blockType == "object") {
        newItem = $("#" + defaultID);
        newItem.html('<input id="name-' + defaultID + '" placeholder="Write the field name and press enter..."></input>');
        field = $("#name-" + defaultID);

        // Show a notification with some information for newbies
        if (!document.cookie.replace(/(?:(?:^|.*;\s*)placeholdertip\s*\=\s*([^;]*).*$)|^.*$/, "$1")) {
          var date = new Date();
          date.setDate(date.getDate() + 365);
          document.cookie = 'placeholdertip=true; expires=' + date.toUTCString + '; path=/';

          notification({
            text: 'Write the field name and then press enter. If you want to create an array or an object, end the name with ":array" or ":object".',
            type: 'information'
          });
        }

        $(field).keypress(function(event) {
          // When you press enter within the new name field:
          if (event.which == 13) {
            event.preventDefault();
            // This var should have a value of the type "name[:array, :object]"
            value = field.val();

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

            if (elements.length == 2 && elements[1] != "array" && elements[1] != "object") {
              notification({
                text: "Only arrays and objects are allowed.",
                type: 'error'
              });
              return false;
            }

            field.remove();

            // TODO: continue here. :) 04/02/2016 10:30pm

            if (typeof blockID === "undefined") {
              blockID = elements[0];
            } else {
              blockID = blockID + '[' + elements[0] + ']';
            }

            if (elements.length == 1) {
              newItem.attr('id', 'block-' + blockID);
              newItem.append('<input name="' + blockID + ':auto" id="' + blockID + '"></input><br>');
              newItem.prepend('<label for="' + blockID + '">' + value + '</label> <span class="actions"><button class="delete">&#8722;</button></span>');
            } else {
              type = "";

              if (elements[1] == "array") {
                type = "array";
              } else {
                type = "object"
              }

              template = "<fieldset id=\"${blockID}\" data-type=\"${type}\"> <h3>${elements[0]}</h3> <span class=\"actions\"> <button class=\"add\">&#43;</button> <button class=\"delete\">&#8722;</button> </span> </fieldset>"
              template = template.replace("${blockID}", blockID);
              template = template.replace("${elements[0]}", elements[0]);
              template = template.replace("${type}", type);
              newItem.after(template);
              newItem.remove();

              console.log('"' + blockID + '" block of type "' + type + '" added.');
            }

            return false;
          }
        });
      }

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
