ace.define("ace/snippets/velocity",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# macro\n\
snippet #macro\n\
	#macro ( ${1:macroName} ${2:\\$var1, [\\$var2, ...]} )\n\
		${3:## macro code}\n\
	#end\n\
# foreach\n\
snippet #foreach\n\
	#foreach ( ${1:\\$item} in ${2:\\$collection} )\n\
		${3:## foreach code}\n\
	#end\n\
# if\n\
snippet #if\n\
	#if ( ${1:true} )\n\
		${0}\n\
	#end\n\
# if ... else\n\
snippet #ife\n\
	#if ( ${1:true} )\n\
		${2}\n\
	#else\n\
		${0}\n\
	#end\n\
#import\n\
snippet #import\n\
	#import ( \"${1:path/to/velocity/format}\" )\n\
# set\n\
snippet #set\n\
	#set ( $${1:var} = ${0} )\n\
";
exports.scope = "velocity";
exports.includeScopes = ["html", "javascript", "css"];

});
