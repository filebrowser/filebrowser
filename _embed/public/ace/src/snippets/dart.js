define("ace/snippets/dart",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet lib\n\
	library ${1};\n\
	${2}\n\
snippet im\n\
	import '${1}';\n\
	${2}\n\
snippet pa\n\
	part '${1}';\n\
	${2}\n\
snippet pao\n\
	part of ${1};\n\
	${2}\n\
snippet main\n\
	void main() {\n\
	  ${1:/* code */}\n\
	}\n\
snippet st\n\
	static ${1}\n\
snippet fi\n\
	final ${1}\n\
snippet re\n\
	return ${1}\n\
snippet br\n\
	break;\n\
snippet th\n\
	throw ${1}\n\
snippet cl\n\
	class ${1:`Filename(\"\", \"untitled\")`} ${2}\n\
snippet imp\n\
	implements ${1}\n\
snippet ext\n\
	extends ${1}\n\
snippet if\n\
	if (${1:true}) {\n\
	  ${2}\n\
	}\n\
snippet ife\n\
	if (${1:true}) {\n\
	  ${2}\n\
	} else {\n\
	  ${3}\n\
	}\n\
snippet el\n\
	else\n\
snippet sw\n\
	switch (${1}) {\n\
	  ${2}\n\
	}\n\
snippet cs\n\
	case ${1}:\n\
	  ${2}\n\
snippet de\n\
	default:\n\
	  ${1}\n\
snippet for\n\
	for (var ${2:i} = 0, len = ${1:things}.length; $2 < len; ${3:++}$2) {\n\
	  ${4:$1[$2]}\n\
	}\n\
snippet fore\n\
	for (final ${2:item} in ${1:itemList}) {\n\
	  ${3:/* code */}\n\
	}\n\
snippet wh\n\
	while (${1:/* condition */}) {\n\
	  ${2:/* code */}\n\
	}\n\
snippet dowh\n\
	do {\n\
	  ${2:/* code */}\n\
	} while (${1:/* condition */});\n\
snippet as\n\
	assert(${1:/* condition */});\n\
snippet try\n\
	try {\n\
	  ${2}\n\
	} catch (${1:Exception e}) {\n\
	}\n\
snippet tryf\n\
	try {\n\
	  ${2}\n\
	} catch (${1:Exception e}) {\n\
	} finally {\n\
	}\n\
";
exports.scope = "dart";

});
