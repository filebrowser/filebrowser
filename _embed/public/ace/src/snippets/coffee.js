define("ace/snippets/coffee",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# Closure loop\n\
snippet forindo\n\
	for ${1:name} in ${2:array}\n\
		do ($1) ->\n\
			${3:// body}\n\
# Array comprehension\n\
snippet fora\n\
	for ${1:name} in ${2:array}\n\
		${3:// body...}\n\
# Object comprehension\n\
snippet foro\n\
	for ${1:key}, ${2:value} of ${3:object}\n\
		${4:// body...}\n\
# Range comprehension (inclusive)\n\
snippet forr\n\
	for ${1:name} in [${2:start}..${3:finish}]\n\
		${4:// body...}\n\
snippet forrb\n\
	for ${1:name} in [${2:start}..${3:finish}] by ${4:step}\n\
		${5:// body...}\n\
# Range comprehension (exclusive)\n\
snippet forrex\n\
	for ${1:name} in [${2:start}...${3:finish}]\n\
		${4:// body...}\n\
snippet forrexb\n\
	for ${1:name} in [${2:start}...${3:finish}] by ${4:step}\n\
		${5:// body...}\n\
# Function\n\
snippet fun\n\
	(${1:args}) ->\n\
		${2:// body...}\n\
# Function (bound)\n\
snippet bfun\n\
	(${1:args}) =>\n\
		${2:// body...}\n\
# Class\n\
snippet cla class ..\n\
	class ${1:`substitute(Filename(), '\\(_\\|^\\)\\(.\\)', '\\u\\2', 'g')`}\n\
		${2}\n\
snippet cla class .. constructor: ..\n\
	class ${1:`substitute(Filename(), '\\(_\\|^\\)\\(.\\)', '\\u\\2', 'g')`}\n\
		constructor: (${2:args}) ->\n\
			${3}\n\
\n\
		${4}\n\
snippet cla class .. extends ..\n\
	class ${1:`substitute(Filename(), '\\(_\\|^\\)\\(.\\)', '\\u\\2', 'g')`} extends ${2:ParentClass}\n\
		${3}\n\
snippet cla class .. extends .. constructor: ..\n\
	class ${1:`substitute(Filename(), '\\(_\\|^\\)\\(.\\)', '\\u\\2', 'g')`} extends ${2:ParentClass}\n\
		constructor: (${3:args}) ->\n\
			${4}\n\
\n\
		${5}\n\
# If\n\
snippet if\n\
	if ${1:condition}\n\
		${2:// body...}\n\
# If __ Else\n\
snippet ife\n\
	if ${1:condition}\n\
		${2:// body...}\n\
	else\n\
		${3:// body...}\n\
# Else if\n\
snippet elif\n\
	else if ${1:condition}\n\
		${2:// body...}\n\
# Ternary If\n\
snippet ifte\n\
	if ${1:condition} then ${2:value} else ${3:other}\n\
# Unless\n\
snippet unl\n\
	${1:action} unless ${2:condition}\n\
# Switch\n\
snippet swi\n\
	switch ${1:object}\n\
		when ${2:value}\n\
			${3:// body...}\n\
\n\
# Log\n\
snippet log\n\
	console.log ${1}\n\
# Try __ Catch\n\
snippet try\n\
	try\n\
		${1}\n\
	catch ${2:error}\n\
		${3}\n\
# Require\n\
snippet req\n\
	${2:$1} = require '${1:sys}'${3}\n\
# Export\n\
snippet exp\n\
	${1:root} = exports ? this\n\
";
exports.scope = "coffee";

});
