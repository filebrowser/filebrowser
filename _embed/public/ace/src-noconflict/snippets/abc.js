ace.define("ace/snippets/abc",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "\n\
snippet zupfnoter.print\n\
	%%%%hn.print {\"startpos\": ${1:pos_y}, \"t\":\"${2:title}\", \"v\":[${3:voices}], \"s\":[[${4:syncvoices}1,2]], \"f\":[${5:flowlines}],  \"sf\":[${6:subflowlines}], \"j\":[${7:jumplines}]}\n\
\n\
snippet zupfnoter.note\n\
	%%%%hn.note {\"pos\": [${1:pos_x},${2:pos_y}], \"text\": \"${3:text}\", \"style\": \"${4:style}\"}\n\
\n\
snippet zupfnoter.annotation\n\
	%%%%hn.annotation {\"id\": \"${1:id}\", \"pos\": [${2:pos}], \"text\": \"${3:text}\"}\n\
\n\
snippet zupfnoter.lyrics\n\
	%%%%hn.lyrics {\"pos\": [${1:x_pos},${2:y_pos}]}\n\
\n\
snippet zupfnoter.legend\n\
	%%%%hn.legend {\"pos\": [${1:x_pos},${2:y_pos}]}\n\
\n\
\n\
\n\
snippet zupfnoter.target\n\
	\"^:${1:target}\"\n\
\n\
snippet zupfnoter.goto\n\
	\"^@${1:target}@${2:distance}\"\n\
\n\
snippet zupfnoter.annotationref\n\
	\"^#${1:target}\"\n\
\n\
snippet zupfnoter.annotation\n\
	\"^!${1:text}@${2:x_offset},${3:y_offset}\"\n\
\n\
\n\
";
exports.scope = "abc";

});
