define("ace/snippets/wollok",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "##\n\
## Basic Java packages and import\n\
snippet im\n\
	import\n\
snippet w.l\n\
	wollok.lang\n\
snippet w.i\n\
	wollok.lib\n\
\n\
## Class and object\n\
snippet cl\n\
	class ${1:`Filename(\"\", \"untitled\")`} ${2}\n\
snippet obj\n\
	object ${1:`Filename(\"\", \"untitled\")`} ${2:inherits Parent}${3}\n\
snippet te\n\
	test ${1:`Filename(\"\", \"untitled\")`}\n\
\n\
##\n\
## Enhancements\n\
snippet inh\n\
	inherits\n\
\n\
##\n\
## Comments\n\
snippet /*\n\
	/*\n\
	 * ${1}\n\
	 */\n\
\n\
##\n\
## Control Statements\n\
snippet el\n\
	else\n\
snippet if\n\
	if (${1}) ${2}\n\
\n\
##\n\
## Create a Method\n\
snippet m\n\
	method ${1:method}(${2}) ${5}\n\
\n\
##  \n\
## Tests\n\
snippet as\n\
	assert.equals(${1:expected}, ${2:actual})\n\
\n\
##\n\
## Exceptions\n\
snippet ca\n\
	catch ${1:e} : (${2:Exception} ) ${3}\n\
snippet thr\n\
	throw\n\
snippet try\n\
	try {\n\
		${3}\n\
	} catch ${1:e} : ${2:Exception} {\n\
	}\n\
\n\
##\n\
## Javadocs\n\
snippet /**\n\
	/**\n\
	 * ${1}\n\
	 */\n\
\n\
##\n\
## Print Methods\n\
snippet print\n\
	console.println(\"${1:Message}\")\n\
\n\
##\n\
## Setter and Getter Methods\n\
snippet set\n\
	method set${1:}(${2:}) {\n\
		$1 = $2\n\
	}\n\
snippet get\n\
	method get${1:}() {\n\
		return ${1:};\n\
	}\n\
\n\
##\n\
## Terminate Methods or Loops\n\
snippet re\n\
	return";
exports.scope = "wollok";

});
