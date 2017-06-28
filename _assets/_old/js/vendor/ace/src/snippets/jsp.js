define("ace/snippets/jsp",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet @page\n\
	<%@page contentType=\"text/html\" pageEncoding=\"UTF-8\"%>\n\
snippet jstl\n\
	<%@ taglib uri=\"http://java.sun.com/jsp/jstl/core\" prefix=\"c\" %>\n\
	<%@ taglib uri=\"http://java.sun.com/jsp/jstl/functions\" prefix=\"fn\" %>\n\
snippet jstl:c\n\
	<%@ taglib uri=\"http://java.sun.com/jsp/jstl/core\" prefix=\"c\" %>\n\
snippet jstl:fn\n\
	<%@ taglib uri=\"http://java.sun.com/jsp/jstl/functions\" prefix=\"fn\" %>\n\
snippet cpath\n\
	${pageContext.request.contextPath}\n\
snippet cout\n\
	<c:out value=\"${1}\" default=\"${2}\" />\n\
snippet cset\n\
	<c:set var=\"${1}\" value=\"${2}\" />\n\
snippet cremove\n\
	<c:remove var=\"${1}\" scope=\"${2:page}\" />\n\
snippet ccatch\n\
	<c:catch var=\"${1}\" />\n\
snippet cif\n\
	<c:if test=\"${${1}}\">\n\
		${2}\n\
	</c:if>\n\
snippet cchoose\n\
	<c:choose>\n\
		${1}\n\
	</c:choose>\n\
snippet cwhen\n\
	<c:when test=\"${${1}}\">\n\
		${2}\n\
	</c:when>\n\
snippet cother\n\
	<c:otherwise>\n\
		${1}\n\
	</c:otherwise>\n\
snippet cfore\n\
	<c:forEach items=\"${${1}}\" var=\"${2}\" varStatus=\"${3}\">\n\
		${4:<c:out value=\"$2\" />}\n\
	</c:forEach>\n\
snippet cfort\n\
	<c:set var=\"${1}\">${2:item1,item2,item3}</c:set>\n\
	<c:forTokens var=\"${3}\" items=\"${$1}\" delims=\"${4:,}\">\n\
		${5:<c:out value=\"$3\" />}\n\
	</c:forTokens>\n\
snippet cparam\n\
	<c:param name=\"${1}\" value=\"${2}\" />\n\
snippet cparam+\n\
	<c:param name=\"${1}\" value=\"${2}\" />\n\
	cparam+${3}\n\
snippet cimport\n\
	<c:import url=\"${1}\" />\n\
snippet cimport+\n\
	<c:import url=\"${1}\">\n\
		<c:param name=\"${2}\" value=\"${3}\" />\n\
		cparam+${4}\n\
	</c:import>\n\
snippet curl\n\
	<c:url value=\"${1}\" var=\"${2}\" />\n\
	<a href=\"${$2}\">${3}</a>\n\
snippet curl+\n\
	<c:url value=\"${1}\" var=\"${2}\">\n\
		<c:param name=\"${4}\" value=\"${5}\" />\n\
		cparam+${6}\n\
	</c:url>\n\
	<a href=\"${$2}\">${3}</a>\n\
snippet credirect\n\
	<c:redirect url=\"${1}\" />\n\
snippet contains\n\
	${fn:contains(${1:string}, ${2:substr})}\n\
snippet contains:i\n\
	${fn:containsIgnoreCase(${1:string}, ${2:substr})}\n\
snippet endswith\n\
	${fn:endsWith(${1:string}, ${2:suffix})}\n\
snippet escape\n\
	${fn:escapeXml(${1:string})}\n\
snippet indexof\n\
	${fn:indexOf(${1:string}, ${2:substr})}\n\
snippet join\n\
	${fn:join(${1:collection}, ${2:delims})}\n\
snippet length\n\
	${fn:length(${1:collection_or_string})}\n\
snippet replace\n\
	${fn:replace(${1:string}, ${2:substr}, ${3:replace})}\n\
snippet split\n\
	${fn:split(${1:string}, ${2:delims})}\n\
snippet startswith\n\
	${fn:startsWith(${1:string}, ${2:prefix})}\n\
snippet substr\n\
	${fn:substring(${1:string}, ${2:begin}, ${3:end})}\n\
snippet substr:a\n\
	${fn:substringAfter(${1:string}, ${2:substr})}\n\
snippet substr:b\n\
	${fn:substringBefore(${1:string}, ${2:substr})}\n\
snippet lc\n\
	${fn:toLowerCase(${1:string})}\n\
snippet uc\n\
	${fn:toUpperCase(${1:string})}\n\
snippet trim\n\
	${fn:trim(${1:string})}\n\
";
exports.scope = "jsp";

});
