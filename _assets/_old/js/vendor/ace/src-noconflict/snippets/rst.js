ace.define("ace/snippets/rst",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# rst\n\
\n\
snippet :\n\
	:${1:field name}: ${2:field body}\n\
snippet *\n\
	*${1:Emphasis}*\n\
snippet **\n\
	**${1:Strong emphasis}**\n\
snippet _\n\
	\\`${1:hyperlink-name}\\`_\n\
	.. _\\`$1\\`: ${2:link-block}\n\
snippet =\n\
	${1:Title}\n\
	=====${2:=}\n\
	${3}\n\
snippet -\n\
	${1:Title}\n\
	-----${2:-}\n\
	${3}\n\
snippet cont:\n\
	.. contents::\n\
	\n\
";
exports.scope = "rst";

});
