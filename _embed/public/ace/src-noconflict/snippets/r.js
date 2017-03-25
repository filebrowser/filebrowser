ace.define("ace/snippets/r",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet #!\n\
	#!/usr/bin/env Rscript\n\
\n\
# includes\n\
snippet lib\n\
	library(${1:package})\n\
snippet req\n\
	require(${1:package})\n\
snippet source\n\
	source('${1:file}')\n\
\n\
# conditionals\n\
snippet if\n\
	if (${1:condition}) {\n\
		${2:code}\n\
	}\n\
snippet el\n\
	else {\n\
		${1:code}\n\
	}\n\
snippet ei\n\
	else if (${1:condition}) {\n\
		${2:code}\n\
	}\n\
\n\
# functions\n\
snippet fun\n\
	${1:name} = function (${2:variables}) {\n\
		${3:code}\n\
	}\n\
snippet ret\n\
	return(${1:code})\n\
\n\
# dataframes, lists, etc\n\
snippet df\n\
	${1:name}[${2:rows}, ${3:cols}]\n\
snippet c\n\
	c(${1:items})\n\
snippet li\n\
	list(${1:items})\n\
snippet mat\n\
	matrix(${1:data}, nrow=${2:rows}, ncol=${3:cols})\n\
\n\
# apply functions\n\
snippet apply\n\
	apply(${1:array}, ${2:margin}, ${3:function})\n\
snippet lapply\n\
	lapply(${1:list}, ${2:function})\n\
snippet sapply\n\
	sapply(${1:list}, ${2:function})\n\
snippet vapply\n\
	vapply(${1:list}, ${2:function}, ${3:type})\n\
snippet mapply\n\
	mapply(${1:function}, ${2:...})\n\
snippet tapply\n\
	tapply(${1:vector}, ${2:index}, ${3:function})\n\
snippet rapply\n\
	rapply(${1:list}, ${2:function})\n\
\n\
# plyr functions\n\
snippet dd\n\
	ddply(${1:frame}, ${2:variables}, ${3:function})\n\
snippet dl\n\
	dlply(${1:frame}, ${2:variables}, ${3:function})\n\
snippet da\n\
	daply(${1:frame}, ${2:variables}, ${3:function})\n\
snippet d_\n\
	d_ply(${1:frame}, ${2:variables}, ${3:function})\n\
\n\
snippet ad\n\
	adply(${1:array}, ${2:margin}, ${3:function})\n\
snippet al\n\
	alply(${1:array}, ${2:margin}, ${3:function})\n\
snippet aa\n\
	aaply(${1:array}, ${2:margin}, ${3:function})\n\
snippet a_\n\
	a_ply(${1:array}, ${2:margin}, ${3:function})\n\
\n\
snippet ld\n\
	ldply(${1:list}, ${2:function})\n\
snippet ll\n\
	llply(${1:list}, ${2:function})\n\
snippet la\n\
	laply(${1:list}, ${2:function})\n\
snippet l_\n\
	l_ply(${1:list}, ${2:function})\n\
\n\
snippet md\n\
	mdply(${1:matrix}, ${2:function})\n\
snippet ml\n\
	mlply(${1:matrix}, ${2:function})\n\
snippet ma\n\
	maply(${1:matrix}, ${2:function})\n\
snippet m_\n\
	m_ply(${1:matrix}, ${2:function})\n\
\n\
# plot functions\n\
snippet pl\n\
	plot(${1:x}, ${2:y})\n\
snippet ggp\n\
	ggplot(${1:data}, aes(${2:aesthetics}))\n\
snippet img\n\
	${1:(jpeg,bmp,png,tiff)}(filename=\"${2:filename}\", width=${3}, height=${4}, unit=\"${5}\")\n\
	${6:plot}\n\
	dev.off()\n\
\n\
# statistical test functions\n\
snippet fis\n\
	fisher.test(${1:x}, ${2:y})\n\
snippet chi\n\
	chisq.test(${1:x}, ${2:y})\n\
snippet tt\n\
	t.test(${1:x}, ${2:y})\n\
snippet wil\n\
	wilcox.test(${1:x}, ${2:y})\n\
snippet cor\n\
	cor.test(${1:x}, ${2:y})\n\
snippet fte\n\
	var.test(${1:x}, ${2:y})\n\
snippet kvt \n\
	kv.test(${1:x}, ${2:y})\n\
";
exports.scope = "r";

});
