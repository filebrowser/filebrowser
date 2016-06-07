var basePath = "/";

$(document).ready(function() {
  basePath += window.location.pathname.split('/')[0];

  // Log out the user sending bad credentials to the server
  $("#logout").click(function(event) {
    event.preventDefault();
    $.ajax({
      type: "GET",
      url: basePath + "",
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


  $(document).pjax('a[data-pjax]', '#content');
});

$(document).on('ready pjax:end', function() {
  $('#content').off();

  // Update the title
  document.title = document.getElementById('site-title').innerHTML;

  //TODO: navbar titles changing effect when changing page

  // Auto Grow Textarea
  function autoGrow() {
    this.style.height = '5px';
    this.style.height = this.scrollHeight + 'px';
  }

  $("textarea").each(autoGrow);
  $('textarea').keyup(autoGrow);
  $(window).resize(function() {
    $("textarea").each(autoGrow);
  });

  if ($('main').hasClass('browse')) {
    $(document).trigger("page:browse");
  }

  if ($(".editor")[0]) {
    $(document).trigger("page:editor");
  }

  return false;
});
