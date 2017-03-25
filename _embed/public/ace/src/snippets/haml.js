define("ace/snippets/haml",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet t\n\
	%table\n\
		%tr\n\
			%th\n\
				${1:headers}\n\
		%tr\n\
			%td\n\
				${2:headers}\n\
snippet ul\n\
	%ul\n\
		%li\n\
			${1:item}\n\
		%li\n\
snippet =rp\n\
	= render :partial => '${1:partial}'\n\
snippet =rpl\n\
	= render :partial => '${1:partial}', :locals => {}\n\
snippet =rpc\n\
	= render :partial => '${1:partial}', :collection => @$1\n\
\n\
";
exports.scope = "haml";

});
