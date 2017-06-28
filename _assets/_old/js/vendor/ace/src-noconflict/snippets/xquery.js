ace.define("ace/snippets/xquery",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet for\n\
	for $${1:item} in ${2:expr}\n\
snippet return\n\
	return ${1:expr}\n\
snippet import\n\
	import module namespace ${1:ns} = \"${2:http://www.example.com/}\";\n\
snippet some\n\
	some $${1:varname} in ${2:expr} satisfies ${3:expr}\n\
snippet every\n\
	every $${1:varname} in ${2:expr} satisfies ${3:expr}\n\
snippet if\n\
	if(${1:true}) then ${2:expr} else ${3:true}\n\
snippet switch\n\
	switch(${1:\"foo\"})\n\
	case ${2:\"foo\"}\n\
	return ${3:true}\n\
	default return ${4:false}\n\
snippet try\n\
	try { ${1:expr} } catch ${2:*} { ${3:expr} }\n\
snippet tumbling\n\
	for tumbling window $${1:varname} in ${2:expr}\n\
	start at $${3:start} when ${4:expr}\n\
	end at $${5:end} when ${6:expr}\n\
	return ${7:expr}\n\
snippet sliding\n\
	for sliding window $${1:varname} in ${2:expr}\n\
	start at $${3:start} when ${4:expr}\n\
	end at $${5:end} when ${6:expr}\n\
	return ${7:expr}\n\
snippet let\n\
	let $${1:varname} := ${2:expr}\n\
snippet group\n\
	group by $${1:varname} := ${2:expr}\n\
snippet order\n\
	order by ${1:expr} ${2:descending}\n\
snippet stable\n\
	stable order by ${1:expr}\n\
snippet count\n\
	count $${1:varname}\n\
snippet ordered\n\
	ordered { ${1:expr} }\n\
snippet unordered\n\
	unordered { ${1:expr} }\n\
snippet treat \n\
	treat as ${1:expr}\n\
snippet castable\n\
	castable as ${1:atomicType}\n\
snippet cast\n\
	cast as ${1:atomicType}\n\
snippet typeswitch\n\
	typeswitch(${1:expr})\n\
	case ${2:type}  return ${3:expr}\n\
	default return ${4:expr}\n\
snippet var\n\
	declare variable $${1:varname} := ${2:expr};\n\
snippet fn\n\
	declare function ${1:ns}:${2:name}(){\n\
	${3:expr}\n\
	};\n\
snippet module\n\
	module namespace ${1:ns} = \"${2:http://www.example.com}\";\n\
";
exports.scope = "xquery";

});
