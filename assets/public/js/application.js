'use strict';

document.addEventListener('DOMContentLoaded', event => {
	document.getElementById('logout').insertAdjacentHTML('beforebegin', `<a href="/admin/settings/">
      <div class="action">
       <i class="material-icons">settings</i>
      </div>
     </a>`);
});


document.addEventListener('listing', event => {
	document.getElementById('newdir').placeholder = "file[:archetype]...";
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
