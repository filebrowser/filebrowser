ace.define("ace/snippets/makefile",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet ifeq\n\
	ifeq (${1:cond0},${2:cond1})\n\
		${3:code}\n\
	endif\n\
";
exports.scope = "makefile";

});
