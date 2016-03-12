$(document).on('page:editor', function() {
  var container = $('.editor');
  var preview = $('#editor-preview');
  var editor = $('#editor-source');

  if (container.hasClass('complete')) {
    // Change title field when editing the header
    $('#content').on('keyup', '#site-title', function() {
      $('.frontmatter #title').val($(this).val());
    });
  }

  if (!container.hasClass('frontmatter-only')) {
    // Setup ace editor
    var mode = $("#editor-source").data('mode');
    var textarea = $('textarea[name="content"]').hide();
    var aceEditor = ace.edit('editor-source');
    aceEditor.getSession().setMode("ace/mode/" + mode);
    aceEditor.getSession().setValue(textarea.val());
    aceEditor.getSession().on('change', function() {
      textarea.val(aceEditor.getSession().getValue());
    });
    aceEditor.setOptions({
      wrap: true,
      maxLines: Infinity,
      theme: "ace/theme/github",
      showPrintMargin: false,
      fontSize: "1em",
      minLines: 20
    });

    $('#content').on('click', '#see-source', function(event) {
      event.preventDefault();
      preview.hide();
      editor.fadeIn();
      $(this).addClass('active');
      $("#see-preview").removeClass('active');
      $("#see-preview").data("previewing", "false");
    })

    // Toggles between preview and editing mode
    $('#content').on('click', '#see-preview', function(event) {
      event.preventDefault();

      // If it currently in the preview mode, hide the preview
      // and show the editor
      if ($(this).data("previewing") == "true") {
        preview.hide();
        editor.fadeIn();
        $(this).removeClass('active');
        $("#see-source").addClass('active');
        $(this).data("previewing", "false");
      } else {
        // If it's in editing mode, convert the markdown to html
        // and show it
        var converter = new showdown.Converter(),
          text = aceEditor.getValue(),
          html = converter.makeHtml(text);

        // Hide the editor and show the preview
        editor.hide();
        preview.html(html).fadeIn();
        $(this).addClass('active');
        $("#see-source").removeClass('active');
        $(this).data("previewing", "true");
      }

      return false;
    });
  }

  $('#content').on('keypress', 'input', function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      $('input[value="Save"]').focus().click();
      return false;
    }
  });

  // Submites any form in the page in JSON format
  $('#content').on('submit', 'form', function(event) {
    event.preventDefault();

    if (!container.hasClass('frontmatter-only')) {
      // Reset preview area and button to make sure it will
      // not serialize any form inside the preview
      preview.html('').fadeOut();
      $("#see-preview").data("previewing", "false");
      editor.fadeIn();
    }

    var button = $(this).find("input[type=submit]:focus");
    var data = {
      content: $(this).serializeJSON(),
      contentType: button.data("type"),
      schedule: button.data("schedule"),
      regenerate: button.data("regenerate")
    }

    var request = new XMLHttpRequest();
    request.open("POST", window.location);
    request.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    request.send(JSON.stringify(data));
    request.onreadystatechange = function() {
      if (request.readyState == 4) {
        var response = JSON.parse(request.responseText),
          type = "success",
          timeout = 5000;

        if (request.status == 200) {
          response.message = button.data("message");
        }

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

  // Adds one more field to the current group
  $("#content").on('click', '.add', function(event) {
    event.preventDefault();
    defaultID = "lorem-ipsum-sin-dolor-amet";

    // Remove if there is an incomplete new item
    newItem = $("#" + defaultID);
    if (newItem.length) {
      newItem.remove();
    }

    block = $(this).parent().parent();
    blockType = block.data("type");
    blockID = block.attr("id");

    // If the Block Type is an array
    if (blockType == "array") {
      newID = blockID + "[]";
      input = blockID;
      input = input.replace(/\[/, '\\[');
      input = input.replace(/\]/, '\\]');
      block.append('<div id="' + newID + '-' + $('#' + input + ' > div').length + '" data-type="array-item"><input name="' + newID + ':auto" id="' + newID + '"></input><span class="actions"> <button class="delete">&#8722;</button></span></div></div>');
      console.log('New array item added.');
    }

    // Main add button, after all blocks
    if (block.is('div') && block.hasClass("frontmatter")) {
      block = $('.blocks');
      blockType = "object";
    }

    // If the Block is an object
    if (blockType == "object") {
      block.append('<div class="block" id="' + defaultID + '"></div>');

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

  $("#content").on('click', '.delete', function(event) {
    event.preventDefault();
    button = $(this);

    name = button.parent().parent().attr("for") || button.parent().parent().attr("id") || button.parent().parent().parent().attr("id");
    name = name.replace(/\[/, '\\[');
    name = name.replace(/\]/, '\\]');
    console.log(name)

    $('label[for="' + name + '"]').fadeOut().remove();
    $('#' + name).fadeOut().remove();

    return false;
  });
});
