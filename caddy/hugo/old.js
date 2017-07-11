
hugo.schedule = function (event) {
  event.preventDefault();

  let date = document.getElementById('date').value;
  if(document.getElementById('publishDate')) {
    date = document.getElementById('publishDate').value;
  }

  buttons.setLoading('publish');

  let data = JSON.stringify(form2js(document.querySelector('form'))),
    headers = {
      'Kind': document.getElementById('editor').dataset.kind,
      'Schedule': 'true'
    };

  webdav.put(window.location.pathname, data, headers)
    .then(() => {
      buttons.setDone('publish');
    })
    .catch(e => {
      console.log(e);
      buttons.setDone('publish', false)
    })
}
