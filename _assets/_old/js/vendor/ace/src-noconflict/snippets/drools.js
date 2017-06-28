ace.define("ace/snippets/drools",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "\n\
snippet rule\n\
	rule \"${1?:rule_name}\"\n\
	when\n\
		${2:// when...} \n\
	then\n\
		${3:// then...}\n\
	end\n\
\n\
snippet query\n\
	query ${1?:query_name}\n\
		${2:// find} \n\
	end\n\
	\n\
snippet declare\n\
	declare ${1?:type_name}\n\
		${2:// attributes} \n\
	end\n\
\n\
";
exports.scope = "drools";

});
