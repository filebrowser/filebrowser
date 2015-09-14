$(document).ready(function() {
  $('form').submit(function(event) {
    var data = JSON.stringify($(this).serializeField())
    var url = $(this).attr('action')

    $.ajax({
      type: 'POST',
      url: url,
      data: data,
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
    $(this).find("> *").each(function() {
      var $this = $(this);
      var name = $this.attr("name");

      if ($this.is("fieldset") && name) {
        result[name] = $this.serializeField();
      } else {
        $.each($this.serializeArray(), function() {
          result[this.name] = this.value;
        });
      }
    });
  });
  return result;
};