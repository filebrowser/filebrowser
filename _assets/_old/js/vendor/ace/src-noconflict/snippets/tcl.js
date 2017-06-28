ace.define("ace/snippets/tcl",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# #!/usr/bin/env tclsh\n\
snippet #!\n\
	#!/usr/bin/env tclsh\n\
	\n\
# Process\n\
snippet pro\n\
	proc ${1:function_name} {${2:args}} {\n\
		${3:#body ...}\n\
	}\n\
#xif\n\
snippet xif\n\
	${1:expr}? ${2:true} : ${3:false}\n\
# Conditional\n\
snippet if\n\
	if {${1}} {\n\
		${2:# body...}\n\
	}\n\
# Conditional if..else\n\
snippet ife\n\
	if {${1}} {\n\
		${2:# body...}\n\
	} else {\n\
		${3:# else...}\n\
	}\n\
# Conditional if..elsif..else\n\
snippet ifee\n\
	if {${1}} {\n\
		${2:# body...}\n\
	} elseif {${3}} {\n\
		${4:# elsif...}\n\
	} else {\n\
		${5:# else...}\n\
	}\n\
# If catch then\n\
snippet ifc\n\
	if { [catch {${1:#do something...}} ${2:err}] } {\n\
		${3:# handle failure...}\n\
	}\n\
# Catch\n\
snippet catch\n\
	catch {${1}} ${2:err} ${3:options}\n\
# While Loop\n\
snippet wh\n\
	while {${1}} {\n\
		${2:# body...}\n\
	}\n\
# For Loop\n\
snippet for\n\
	for {set ${2:var} 0} {$$2 < ${1:count}} {${3:incr} $2} {\n\
		${4:# body...}\n\
	}\n\
# Foreach Loop\n\
snippet fore\n\
	foreach ${1:x} {${2:#list}} {\n\
		${3:# body...}\n\
	}\n\
# after ms script...\n\
snippet af\n\
	after ${1:ms} ${2:#do something}\n\
# after cancel id\n\
snippet afc\n\
	after cancel ${1:id or script}\n\
# after idle\n\
snippet afi\n\
	after idle ${1:script}\n\
# after info id\n\
snippet afin\n\
	after info ${1:id}\n\
# Expr\n\
snippet exp\n\
	expr {${1:#expression here}}\n\
# Switch\n\
snippet sw\n\
	switch ${1:var} {\n\
		${3:pattern 1} {\n\
			${4:#do something}\n\
		}\n\
		default {\n\
			${2:#do something}\n\
		}\n\
	}\n\
# Case\n\
snippet ca\n\
	${1:pattern} {\n\
		${2:#do something}\n\
	}${3}\n\
# Namespace eval\n\
snippet ns\n\
	namespace eval ${1:path} {${2:#script...}}\n\
# Namespace current\n\
snippet nsc\n\
	namespace current\n\
";
exports.scope = "tcl";

});
