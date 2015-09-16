$(document).ready(function() {
  $(document).pjax('a', '#container');
  $('.scroll').perfectScrollbar();

  $("#preview").click(function(e) {
    e.preventDefault();

    var preview = $("#preview-area"),
      editor = $('.editor textarea');

    if ($(this).data("previewing") == "true") {
      preview.hide();
      editor.fadeIn();
      $(this).data("previewing", "false");

      notification({
        text: "You've gone into editing mode.",
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
        text: "You've gone into preview mode.",
        type: 'information',
        timeout: 2000
      });
    }

    return false;
  });

  $('form').submit(function(event) {
    var data = JSON.stringify($(this).serializeForm()),
      url = $(this).attr('action'),
      button = $(this).find("input[type=submit]:focus"),
      action = button.val();

    $.ajax({
      type: 'POST',
      url: url,
      data: data,
      beforeSend: function(xhr) {
        xhr.setRequestHeader('X-Save-Mode', action);
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

    event.preventDefault();
  });

  $("#logout").click(function(e) {
    e.preventDefault();
    jQuery.ajax({
        type: "GET",
        url: "/admin",
        async: false,
        username: "logmeout",
        password: "123456",
        headers: {
          "Authorization": "Basic xxx"
        }
      })
      .fail(function() {
        window.location = "/";
      });
    return false;
  });

  $(".add").click(function(e) {
    e.preventDefault();
    fieldset = $(this).closest("fieldset");
    fieldset.append("<input name=\"" + fieldset.attr("name") + "\" id=\"" + fieldset.attr("name") + "\" value=\"\"></input><br>");
    return false;
  });
});


$.fn.serializeForm = function() {
  var result = {};
  this.each(function() {
    $(this).find(".data > *").each(function() {
      var $this = $(this);
      var name = $this.attr("name");

      if ($this.is("fieldset") && name) {
        if ($this.attr("type") == "array") {
          result[this.name] = [];

          $.each($this.serializeArray(), function() {
            result[this.name].push(this.value);
          });
        } else {
          result[name] = $this.serializeForm();
        }
      } else {
        $.each($this.serializeArray(), function() {
          result[this.name] = this.value;
        });
      }
    });
  });
  return result;
};