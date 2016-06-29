'use strict';

document.addEventListener('DOMContentLoaded', event => {
    document.getElementById('logout').insertAdjacentHTML('beforebegin', `<a href="/admin/settings/">
      <div class="action">
       <i class="material-icons">settings</i>
      </div>
     </a>`);
});

document.addEventListener('listing', event => {
    if (window.location.pathname.includes('/content/')) {
        document.getElementById('newdir').placeholder = "file[:archetype]...";
        document.getElementById('newdir').removeEventListener('keydown', newDirEvent);
        document.getElementById('newdir').addEventListener('keydown', event => {

            if (event.keyCode == 27) {
                document.getElementById('newdir').classList.toggle('enabled');
                setTimeout(() => {
                    document.getElementById('newdir').value = '';
                }, 200);
            }

            if (event.keyCode == 13) {
                event.preventDefault();

                let value = document.getElementById('newdir').value;
                let index = value.lastIndexOf(':');
                let name = value.substring(0, index);
                let archetype = value.substring(index + 1, value.length);
                if (name == "") name = archetype;
                if (index == -1) archetype = "";

                let button = document.getElementById('new');
                let html = button.changeToLoading();
                let request = new XMLHttpRequest();
                request.open("POST", window.location);
                request.setRequestHeader('Filename', name);
                request.setRequestHeader('Archetype', archetype);
                request.send();
                request.onreadystatechange = function() {
                    if (request.readyState == 4) {
                        button.changeToDone((request.status != 200), html);
                        if (request.status == 200) {
                            window.location = window.location.pathname + name;
                        }
                    }
                }
            }
        });
    }
});

document.addEventListener('editor', event => {
    document.getElementById('submit').insertAdjacentHTML('afterend', `<div class="right">
	 <button id="publish">
		 <span>
			<i class="material-icons">send</i>
		 </span>
		 <span>publish</span>
	 </button>
	 </div>`);

    if (document.getElementById('date') || document.getElementById('publishdate')) {
        document.querySelector('#editor .right').insertAdjacentHTML('afterbegin', ` <button id="schedule">
			  <span>
				 <i class="material-icons">alarm</i>
			  </span>
			  <span>Schedule</span>
		  </button>`);

        document.getElementById('schedule').addEventListener('click', event => {
            event.preventDefault();
        });
    }

    document.getElementById('publish').addEventListener('click', event => {
        console.log("Hey")
        event.preventDefault();

        if (document.getElementById('draft')) {
            document.getElementById('block-draft').remove();
        }
        let container = document.getElementById('editor');
        let kind = container.dataset.kind;
        let button = document.querySelector('#publish span:first-child');

        let data = form2js(document.querySelector('form'));
        let html = button.changeToLoading();
        let request = new XMLHttpRequest();
        request.open("PUT", window.location);
        request.setRequestHeader('Kind', kind);
        request.setRequestHeader('Regenerate', "true");
        request.send(JSON.stringify(data));
        request.onreadystatechange = function() {
            if (request.readyState == 4) {
                button.changeToDone((request.status != 200), html);
            }
        }
    });
});