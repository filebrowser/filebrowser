$(document).ready(function() {
  $('.scroll').perfectScrollbar();


  $('form').submit(function(event) {
    var data = JSON.stringify($(this).serializeField())
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
      alert("it workss");
    }).fail(function(data) {
      alert("it failed");
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
});


$.fn.serializeField = function() {
  var result = {};
  this.each(function() {
    $(this).find(".container > *").each(function() {
      var $this = $(this);
      var name = $this.attr("name");

      if ($this.is("fieldset") && name) {
        if ($this.attr("type") == "array") {
          result[this.name] = [];

          $.each($this.serializeArray(), function() {
            result[this.name].push(this.value);
          });
        } else {
          result[name] = $this.serializeField();
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