ace.define("ace/snippets/textile",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# Jekyll post header\n\
snippet header\n\
	---\n\
	title: ${1:title}\n\
	layout: post\n\
	date: ${2:date} ${3:hour:minute:second} -05:00\n\
	---\n\
\n\
# Image\n\
snippet img\n\
	!${1:url}(${2:title}):${3:link}!\n\
\n\
# Table\n\
snippet |\n\
	|${1}|${2}\n\
\n\
# Link\n\
snippet link\n\
	\"${1:link text}\":${2:url}\n\
\n\
# Acronym\n\
snippet (\n\
	(${1:Expand acronym})${2}\n\
\n\
# Footnote\n\
snippet fn\n\
	[${1:ref number}] ${3}\n\
\n\
	fn$1. ${2:footnote}\n\
	\n\
";
exports.scope = "textile";

});
