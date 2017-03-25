ace.define("ace/snippets/haskell",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet lang\n\
	{-# LANGUAGE ${1:OverloadedStrings} #-}\n\
snippet info\n\
	-- |\n\
	-- Module      :  ${1:Module.Namespace}\n\
	-- Copyright   :  ${2:Author} ${3:2011-2012}\n\
	-- License     :  ${4:BSD3}\n\
	--\n\
	-- Maintainer  :  ${5:email@something.com}\n\
	-- Stability   :  ${6:experimental}\n\
	-- Portability :  ${7:unknown}\n\
	--\n\
	-- ${8:Description}\n\
	--\n\
snippet import\n\
	import           ${1:Data.Text}\n\
snippet import2\n\
	import           ${1:Data.Text} (${2:head})\n\
snippet importq\n\
	import qualified ${1:Data.Text} as ${2:T}\n\
snippet inst\n\
	instance ${1:Monoid} ${2:Type} where\n\
		${3}\n\
snippet type\n\
	type ${1:Type} = ${2:Type}\n\
snippet data\n\
	data ${1:Type} = ${2:$1} ${3:Int}\n\
snippet newtype\n\
	newtype ${1:Type} = ${2:$1} ${3:Int}\n\
snippet class\n\
	class ${1:Class} a where\n\
		${2}\n\
snippet module\n\
	module `substitute(substitute(expand('%:r'), '[/\\\\]','.','g'),'^\\%(\\l*\\.\\)\\?','','')` (\n\
	)	where\n\
	`expand('%') =~ 'Main' ? \"\\n\\nmain = do\\n  print \\\"hello world\\\"\" : \"\"`\n\
\n\
snippet const\n\
	${1:name} :: ${2:a}\n\
	$1 = ${3:undefined}\n\
snippet fn\n\
	${1:fn} :: ${2:a} -> ${3:a}\n\
	$1 ${4} = ${5:undefined}\n\
snippet fn2\n\
	${1:fn} :: ${2:a} -> ${3:a} -> ${4:a}\n\
	$1 ${5} = ${6:undefined}\n\
snippet ap\n\
	${1:map} ${2:fn} ${3:list}\n\
snippet do\n\
	do\n\
		\n\
snippet λ\n\
	\\${1:x} -> ${2}\n\
snippet \\\n\
	\\${1:x} -> ${2}\n\
snippet <-\n\
	${1:a} <- ${2:m a}\n\
snippet ←\n\
	${1:a} <- ${2:m a}\n\
snippet ->\n\
	${1:m a} -> ${2:a}\n\
snippet →\n\
	${1:m a} -> ${2:a}\n\
snippet tup\n\
	(${1:a}, ${2:b})\n\
snippet tup2\n\
	(${1:a}, ${2:b}, ${3:c})\n\
snippet tup3\n\
	(${1:a}, ${2:b}, ${3:c}, ${4:d})\n\
snippet rec\n\
	${1:Record} { ${2:recFieldA} = ${3:undefined}\n\
				, ${4:recFieldB} = ${5:undefined}\n\
				}\n\
snippet case\n\
	case ${1:something} of\n\
		${2} -> ${3}\n\
snippet let\n\
	let ${1} = ${2}\n\
	in ${3}\n\
snippet where\n\
	where\n\
		${1:fn} = ${2:undefined}\n\
";
exports.scope = "haskell";

});
