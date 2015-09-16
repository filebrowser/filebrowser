$(document).ready(function() {
  $('.scroll').perfectScrollbar();

  $("#preview").click(function(e) {
    e.preventDefault();

    var preview = $("#preview-area"),
      editor = $('.editor textarea');

    if ($(this).attr("previewing") == "true") {
      preview.hide();
      editor.fadeIn();
      $(this).attr("previewing", "false");
    } else {
      var converter = new showdown.Converter(),
        text = editor.val(),
        html = converter.makeHtml(text);

      editor.hide();
      preview.html(html).fadeIn();
      $(this).attr("previewing", "true");
    }

    return false;
  });

  $('form').submit(function(event) {
    var data = JSON.stringify($(this).serializeForm())
    var url = $(this).attr('action')
    var action = $(this).find("input[type=submit]:focus").val();

    console.log(data);

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
      if (action == "Save") {
        var word = "saved";
      } else {
        var word = "published";
      }

      notification({
        text: 'The post was ' + word + '.',
        type: 'success'
      });
    }).fail(function(data) {
      notification({
        text: 'Something went wrong.',
        type: 'error'
      });
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