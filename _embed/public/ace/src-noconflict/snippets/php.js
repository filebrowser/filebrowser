ace.define("ace/snippets/php",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet <?\n\
	<?php\n\
\n\
	${1}\n\
snippet ec\n\
	echo ${1};\n\
snippet <?e\n\
	<?php echo ${1} ?>\n\
# this one is for php5.4\n\
snippet <?=\n\
	<?=${1}?>\n\
snippet ns\n\
	namespace ${1:Foo\\Bar\\Baz};\n\
	${2}\n\
snippet use\n\
	use ${1:Foo\\Bar\\Baz};\n\
	${2}\n\
snippet c\n\
	${1:abstract }class ${2:$FILENAME}\n\
	{\n\
		${3}\n\
	}\n\
snippet i\n\
	interface ${1:$FILENAME}\n\
	{\n\
		${2}\n\
	}\n\
snippet t.\n\
	$this->${1}\n\
snippet f\n\
	function ${1:foo}(${2:array }${3:$bar})\n\
	{\n\
		${4}\n\
	}\n\
# method\n\
snippet m\n\
	${1:abstract }${2:protected}${3: static} function ${4:foo}(${5:array }${6:$bar})\n\
	{\n\
		${7}\n\
	}\n\
# setter method\n\
snippet sm \n\
	/**\n\
	 * Sets the value of ${1:foo}\n\
	 *\n\
	 * @param ${2:$1} $$1 ${3:description}\n\
	 *\n\
	 * @return ${4:$FILENAME}\n\
	 */\n\
	${5:public} function set${6:$2}(${7:$2 }$$1)\n\
	{\n\
		$this->${8:$1} = $$1;\n\
		return $this;\n\
	}${9}\n\
# getter method\n\
snippet gm\n\
	/**\n\
	 * Gets the value of ${1:foo}\n\
	 *\n\
	 * @return ${2:$1}\n\
	 */\n\
	${3:public} function get${4:$2}()\n\
	{\n\
		return $this->${5:$1};\n\
	}${6}\n\
#setter\n\
snippet $s\n\
	${1:$foo}->set${2:Bar}(${3});\n\
#getter\n\
snippet $g\n\
	${1:$foo}->get${2:Bar}();\n\
\n\
# Tertiary conditional\n\
snippet =?:\n\
	$${1:foo} = ${2:true} ? ${3:a} : ${4};\n\
snippet ?:\n\
	${1:true} ? ${2:a} : ${3}\n\
\n\
snippet C\n\
	$_COOKIE['${1:variable}']${2}\n\
snippet E\n\
	$_ENV['${1:variable}']${2}\n\
snippet F\n\
	$_FILES['${1:variable}']${2}\n\
snippet G\n\
	$_GET['${1:variable}']${2}\n\
snippet P\n\
	$_POST['${1:variable}']${2}\n\
snippet R\n\
	$_REQUEST['${1:variable}']${2}\n\
snippet S\n\
	$_SERVER['${1:variable}']${2}\n\
snippet SS\n\
	$_SESSION['${1:variable}']${2}\n\
	\n\
# the following are old ones\n\
snippet inc\n\
	include '${1:file}';${2}\n\
snippet inc1\n\
	include_once '${1:file}';${2}\n\
snippet req\n\
	require '${1:file}';${2}\n\
snippet req1\n\
	require_once '${1:file}';${2}\n\
# Start Docblock\n\
snippet /*\n\
	/**\n\
	 * ${1}\n\
	 */\n\
# Class - post doc\n\
snippet doc_cp\n\
	/**\n\
	 * ${1:undocumented class}\n\
	 *\n\
	 * @package ${2:default}\n\
	 * @subpackage ${3:default}\n\
	 * @author ${4:`g:snips_author`}\n\
	 */${5}\n\
# Class Variable - post doc\n\
snippet doc_vp\n\
	/**\n\
	 * ${1:undocumented class variable}\n\
	 *\n\
	 * @var ${2:string}\n\
	 */${3}\n\
# Class Variable\n\
snippet doc_v\n\
	/**\n\
	 * ${3:undocumented class variable}\n\
	 *\n\
	 * @var ${4:string}\n\
	 */\n\
	${1:var} $${2};${5}\n\
# Class\n\
snippet doc_c\n\
	/**\n\
	 * ${3:undocumented class}\n\
	 *\n\
	 * @package ${4:default}\n\
	 * @subpackage ${5:default}\n\
	 * @author ${6:`g:snips_author`}\n\
	 */\n\
	${1:}class ${2:}\n\
	{\n\
		${7}\n\
	} // END $1class $2\n\
# Constant Definition - post doc\n\
snippet doc_dp\n\
	/**\n\
	 * ${1:undocumented constant}\n\
	 */${2}\n\
# Constant Definition\n\
snippet doc_d\n\
	/**\n\
	 * ${3:undocumented constant}\n\
	 */\n\
	define(${1}, ${2});${4}\n\
# Function - post doc\n\
snippet doc_fp\n\
	/**\n\
	 * ${1:undocumented function}\n\
	 *\n\
	 * @return ${2:void}\n\
	 * @author ${3:`g:snips_author`}\n\
	 */${4}\n\
# Function signature\n\
snippet doc_s\n\
	/**\n\
	 * ${4:undocumented function}\n\
	 *\n\
	 * @return ${5:void}\n\
	 * @author ${6:`g:snips_author`}\n\
	 */\n\
	${1}function ${2}(${3});${7}\n\
# Function\n\
snippet doc_f\n\
	/**\n\
	 * ${4:undocumented function}\n\
	 *\n\
	 * @return ${5:void}\n\
	 * @author ${6:`g:snips_author`}\n\
	 */\n\
	${1}function ${2}(${3})\n\
	{${7}\n\
	}\n\
# Header\n\
snippet doc_h\n\
	/**\n\
	 * ${1}\n\
	 *\n\
	 * @author ${2:`g:snips_author`}\n\
	 * @version ${3:$Id$}\n\
	 * @copyright ${4:$2}, `strftime('%d %B, %Y')`\n\
	 * @package ${5:default}\n\
	 */\n\
	\n\
# Interface\n\
snippet interface\n\
	/**\n\
	 * ${2:undocumented class}\n\
	 *\n\
	 * @package ${3:default}\n\
	 * @author ${4:`g:snips_author`}\n\
	 */\n\
	interface ${1:$FILENAME}\n\
	{\n\
		${5}\n\
	}\n\
# class ...\n\
snippet class\n\
	/**\n\
	 * ${1}\n\
	 */\n\
	class ${2:$FILENAME}\n\
	{\n\
		${3}\n\
		/**\n\
		 * ${4}\n\
		 */\n\
		${5:public} function ${6:__construct}(${7:argument})\n\
		{\n\
			${8:// code...}\n\
		}\n\
	}\n\
# define(...)\n\
snippet def\n\
	define('${1}'${2});${3}\n\
# defined(...)\n\
snippet def?\n\
	${1}defined('${2}')${3}\n\
snippet wh\n\
	while (${1:/* condition */}) {\n\
		${2:// code...}\n\
	}\n\
# do ... while\n\
snippet do\n\
	do {\n\
		${2:// code... }\n\
	} while (${1:/* condition */});\n\
snippet if\n\
	if (${1:/* condition */}) {\n\
		${2:// code...}\n\
	}\n\
snippet ifil\n\
	<?php if (${1:/* condition */}): ?>\n\
		${2:<!-- code... -->}\n\
	<?php endif; ?>\n\
snippet ife\n\
	if (${1:/* condition */}) {\n\
		${2:// code...}\n\
	} else {\n\
		${3:// code...}\n\
	}\n\
	${4}\n\
snippet ifeil\n\
	<?php if (${1:/* condition */}): ?>\n\
		${2:<!-- html... -->}\n\
	<?php else: ?>\n\
		${3:<!-- html... -->}\n\
	<?php endif; ?>\n\
	${4}\n\
snippet else\n\
	else {\n\
		${1:// code...}\n\
	}\n\
snippet elseif\n\
	elseif (${1:/* condition */}) {\n\
		${2:// code...}\n\
	}\n\
snippet switch\n\
	switch ($${1:variable}) {\n\
		case '${2:value}':\n\
			${3:// code...}\n\
			break;\n\
		${5}\n\
		default:\n\
			${4:// code...}\n\
			break;\n\
	}\n\
snippet case\n\
	case '${1:value}':\n\
		${2:// code...}\n\
		break;${3}\n\
snippet for\n\
	for ($${2:i} = 0; $$2 < ${1:count}; $$2${3:++}) {\n\
		${4: // code...}\n\
	}\n\
snippet foreach\n\
	foreach ($${1:variable} as $${2:value}) {\n\
		${3:// code...}\n\
	}\n\
snippet foreachil\n\
	<?php foreach ($${1:variable} as $${2:value}): ?>\n\
		${3:<!-- html... -->}\n\
	<?php endforeach; ?>\n\
snippet foreachk\n\
	foreach ($${1:variable} as $${2:key} => $${3:value}) {\n\
		${4:// code...}\n\
	}\n\
snippet foreachkil\n\
	<?php foreach ($${1:variable} as $${2:key} => $${3:value}): ?>\n\
		${4:<!-- html... -->}\n\
	<?php endforeach; ?>\n\
# $... = array (...)\n\
snippet array\n\
	$${1:arrayName} = array('${2}' => ${3});${4}\n\
snippet try\n\
	try {\n\
		${2}\n\
	} catch (${1:Exception} $e) {\n\
	}\n\
# lambda with closure\n\
snippet lambda\n\
	${1:static }function (${2:args}) use (${3:&$x, $y /*put vars in scope (closure) */}) {\n\
		${4}\n\
	};\n\
# pre_dump();\n\
snippet pd\n\
	echo '<pre>'; var_dump(${1}); echo '</pre>';\n\
# pre_dump(); die();\n\
snippet pdd\n\
	echo '<pre>'; var_dump(${1}); echo '</pre>'; die(${2:});\n\
snippet vd\n\
	var_dump(${1});\n\
snippet vdd\n\
	var_dump(${1}); die(${2:});\n\
snippet http_redirect\n\
	header (\"HTTP/1.1 301 Moved Permanently\"); \n\
	header (\"Location: \".URL); \n\
	exit();\n\
# Getters & Setters\n\
snippet gs\n\
	/**\n\
	 * Gets the value of ${1:foo}\n\
	 *\n\
	 * @return ${2:$1}\n\
	 */\n\
	public function get${3:$2}()\n\
	{\n\
		return $this->${4:$1};\n\
	}\n\
\n\
	/**\n\
	 * Sets the value of $1\n\
	 *\n\
	 * @param $2 $$1 ${5:description}\n\
	 *\n\
	 * @return ${6:$FILENAME}\n\
	 */\n\
	public function set$3(${7:$2 }$$1)\n\
	{\n\
		$this->$4 = $$1;\n\
		return $this;\n\
	}${8}\n\
# anotation, get, and set, useful for doctrine\n\
snippet ags\n\
	/**\n\
	 * ${1:description}\n\
	 * \n\
	 * @${7}\n\
	 */\n\
	${2:protected} $${3:foo};\n\
\n\
	public function get${4:$3}()\n\
	{\n\
		return $this->$3;\n\
	}\n\
\n\
	public function set$4(${5:$4 }$${6:$3})\n\
	{\n\
		$this->$3 = $$6;\n\
		return $this;\n\
	}\n\
snippet rett\n\
	return true;\n\
snippet retf\n\
	return false;\n\
";
exports.scope = "php";

});
