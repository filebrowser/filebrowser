'use strict';

document.addEventListener('DOMContentLoaded', event => {
	document.getElementById('logout').insertAdjacentHTML('beforebegin', `<a href="/admin/settings/">
      <div class="action">
       <i class="material-icons">settings</i>
      </div>
     </a>`);

	document.getElementById('submit').insertAdjacentHTML('afterend', `<div class="right">
     <button id="publish" type="submit" data-type="content-only">
         <span>
            <i class="material-icons">send</i>
         </span>
         <span>publish</span>
     </button>
     </div>`);

	if (document.getElementById('date') || document.getElementById('publishdate')) {
		document.querySelector('#editor .right').insertAdjacentHTML('afterbegin', ` <button id="schedule" type="submit" data-type="content-only">
              <span>
                 <i class="material-icons">alarm</i>
              </span>
              <span>Schedule</span>
          </button>`);
	}



});
