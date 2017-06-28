define("ace/snippets/javascript",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# Prototype\n\
snippet proto\n\
	${1:class_name}.prototype.${2:method_name} = function(${3:first_argument}) {\n\
		${4:// body...}\n\
	};\n\
# Function\n\
snippet fun\n\
	function ${1?:function_name}(${2:argument}) {\n\
		${3:// body...}\n\
	}\n\
# Anonymous Function\n\
regex /((=)\\s*|(:)\\s*|(\\()|\\b)/f/(\\))?/\n\
snippet f\n\
	function${M1?: ${1:functionName}}($2) {\n\
		${0:$TM_SELECTED_TEXT}\n\
	}${M2?;}${M3?,}${M4?)}\n\
# Immediate function\n\
trigger \\(?f\\(\n\
endTrigger \\)?\n\
snippet f(\n\
	(function(${1}) {\n\
		${0:${TM_SELECTED_TEXT:/* code */}}\n\
	}(${1}));\n\
# if\n\
snippet if\n\
	if (${1:true}) {\n\
		${0}\n\
	}\n\
# if ... else\n\
snippet ife\n\
	if (${1:true}) {\n\
		${2}\n\
	} else {\n\
		${0}\n\
	}\n\
# tertiary conditional\n\
snippet ter\n\
	${1:/* condition */} ? ${2:a} : ${3:b}\n\
# switch\n\
snippet switch\n\
	switch (${1:expression}) {\n\
		case '${3:case}':\n\
			${4:// code}\n\
			break;\n\
		${5}\n\
		default:\n\
			${2:// code}\n\
	}\n\
# case\n\
snippet case\n\
	case '${1:case}':\n\
		${2:// code}\n\
		break;\n\
	${3}\n\
\n\
# while (...) {...}\n\
snippet wh\n\
	while (${1:/* condition */}) {\n\
		${0:/* code */}\n\
	}\n\
# try\n\
snippet try\n\
	try {\n\
		${0:/* code */}\n\
	} catch (e) {}\n\
# do...while\n\
snippet do\n\
	do {\n\
		${2:/* code */}\n\
	} while (${1:/* condition */});\n\
# Object Method\n\
snippet :f\n\
regex /([,{[])|^\\s*/:f/\n\
	${1:method_name}: function(${2:attribute}) {\n\
		${0}\n\
	}${3:,}\n\
# setTimeout function\n\
snippet setTimeout\n\
regex /\\b/st|timeout|setTimeo?u?t?/\n\
	setTimeout(function() {${3:$TM_SELECTED_TEXT}}, ${1:10});\n\
# Get Elements\n\
snippet gett\n\
	getElementsBy${1:TagName}('${2}')${3}\n\
# Get Element\n\
snippet get\n\
	getElementBy${1:Id}('${2}')${3}\n\
# console.log (Firebug)\n\
snippet cl\n\
	console.log(${1});\n\
# return\n\
snippet ret\n\
	return ${1:result}\n\
# for (property in object ) { ... }\n\
snippet fori\n\
	for (var ${1:prop} in ${2:Things}) {\n\
		${0:$2[$1]}\n\
	}\n\
# hasOwnProperty\n\
snippet has\n\
	hasOwnProperty(${1})\n\
# docstring\n\
snippet /**\n\
	/**\n\
	 * ${1:description}\n\
	 *\n\
	 */\n\
snippet @par\n\
regex /^\\s*\\*\\s*/@(para?m?)?/\n\
	@param {${1:type}} ${2:name} ${3:description}\n\
snippet @ret\n\
	@return {${1:type}} ${2:description}\n\
# JSON.parse\n\
snippet jsonp\n\
	JSON.parse(${1:jstr});\n\
# JSON.stringify\n\
snippet jsons\n\
	JSON.stringify(${1:object});\n\
# self-defining function\n\
snippet sdf\n\
	var ${1:function_name} = function(${2:argument}) {\n\
		${3:// initial code ...}\n\
\n\
		$1 = function($2) {\n\
			${4:// main code}\n\
		};\n\
	}\n\
# singleton\n\
snippet sing\n\
	function ${1:Singleton} (${2:argument}) {\n\
		// the cached instance\n\
		var instance;\n\
\n\
		// rewrite the constructor\n\
		$1 = function $1($2) {\n\
			return instance;\n\
		};\n\
		\n\
		// carry over the prototype properties\n\
		$1.prototype = this;\n\
\n\
		// the instance\n\
		instance = new $1();\n\
\n\
		// reset the constructor pointer\n\
		instance.constructor = $1;\n\
\n\
		${3:// code ...}\n\
\n\
		return instance;\n\
	}\n\
# class\n\
snippet class\n\
regex /^\\s*/clas{0,2}/\n\
	var ${1:class} = function(${20}) {\n\
		$40$0\n\
	};\n\
	\n\
	(function() {\n\
		${60:this.prop = \"\"}\n\
	}).call(${1:class}.prototype);\n\
	\n\
	exports.${1:class} = ${1:class};\n\
# \n\
snippet for-\n\
	for (var ${1:i} = ${2:Things}.length; ${1:i}--; ) {\n\
		${0:${2:Things}[${1:i}];}\n\
	}\n\
# for (...) {...}\n\
snippet for\n\
	for (var ${1:i} = 0; $1 < ${2:Things}.length; $1++) {\n\
		${3:$2[$1]}$0\n\
	}\n\
# for (...) {...} (Improved Native For-Loop)\n\
snippet forr\n\
	for (var ${1:i} = ${2:Things}.length - 1; $1 >= 0; $1--) {\n\
		${3:$2[$1]}$0\n\
	}\n\
\n\
\n\
#modules\n\
snippet def\n\
	define(function(require, exports, module) {\n\
	\"use strict\";\n\
	var ${1/.*\\///} = require(\"${1}\");\n\
	\n\
	$TM_SELECTED_TEXT\n\
	});\n\
snippet req\n\
guard ^\\s*\n\
	var ${1/.*\\///} = require(\"${1}\");\n\
	$0\n\
snippet requ\n\
guard ^\\s*\n\
	var ${1/.*\\/(.)/\\u$1/} = require(\"${1}\").${1/.*\\/(.)/\\u$1/};\n\
	$0\n\
";
exports.scope = "javascript";

});
