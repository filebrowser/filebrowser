ace.define("ace/snippets/java",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "## Access Modifiers\n\
snippet po\n\
	protected\n\
snippet pu\n\
	public\n\
snippet pr\n\
	private\n\
##\n\
## Annotations\n\
snippet before\n\
	@Before\n\
	static void ${1:intercept}(${2:args}) { ${3} }\n\
snippet mm\n\
	@ManyToMany\n\
	${1}\n\
snippet mo\n\
	@ManyToOne\n\
	${1}\n\
snippet om\n\
	@OneToMany${1:(cascade=CascadeType.ALL)}\n\
	${2}\n\
snippet oo\n\
	@OneToOne\n\
	${1}\n\
##\n\
## Basic Java packages and import\n\
snippet im\n\
	import\n\
snippet j.b\n\
	java.beans.\n\
snippet j.i\n\
	java.io.\n\
snippet j.m\n\
	java.math.\n\
snippet j.n\n\
	java.net.\n\
snippet j.u\n\
	java.util.\n\
##\n\
## Class\n\
snippet cl\n\
	class ${1:`Filename(\"\", \"untitled\")`} ${2}\n\
snippet in\n\
	interface ${1:`Filename(\"\", \"untitled\")`} ${2:extends Parent}${3}\n\
snippet tc\n\
	public class ${1:`Filename()`} extends ${2:TestCase}\n\
##\n\
## Class Enhancements\n\
snippet ext\n\
	extends \n\
snippet imp\n\
	implements\n\
##\n\
## Comments\n\
snippet /*\n\
	/*\n\
	 * ${1}\n\
	 */\n\
##\n\
## Constants\n\
snippet co\n\
	static public final ${1:String} ${2:var} = ${3};${4}\n\
snippet cos\n\
	static public final String ${1:var} = \"${2}\";${3}\n\
##\n\
## Control Statements\n\
snippet case\n\
	case ${1}:\n\
		${2}\n\
snippet def\n\
	default:\n\
		${2}\n\
snippet el\n\
	else\n\
snippet elif\n\
	else if (${1}) ${2}\n\
snippet if\n\
	if (${1}) ${2}\n\
snippet sw\n\
	switch (${1}) {\n\
		${2}\n\
	}\n\
##\n\
## Create a Method\n\
snippet m\n\
	${1:void} ${2:method}(${3}) ${4:throws }${5}\n\
##\n\
## Create a Variable\n\
snippet v\n\
	${1:String} ${2:var}${3: = null}${4};${5}\n\
##\n\
## Enhancements to Methods, variables, classes, etc.\n\
snippet ab\n\
	abstract\n\
snippet fi\n\
	final\n\
snippet st\n\
	static\n\
snippet sy\n\
	synchronized\n\
##\n\
## Error Methods\n\
snippet err\n\
	System.err.print(\"${1:Message}\");\n\
snippet errf\n\
	System.err.printf(\"${1:Message}\", ${2:exception});\n\
snippet errln\n\
	System.err.println(\"${1:Message}\");\n\
##\n\
## Exception Handling\n\
snippet as\n\
	assert ${1:test} : \"${2:Failure message}\";${3}\n\
snippet ca\n\
	catch(${1:Exception} ${2:e}) ${3}\n\
snippet thr\n\
	throw\n\
snippet ths\n\
	throws\n\
snippet try\n\
	try {\n\
		${3}\n\
	} catch(${1:Exception} ${2:e}) {\n\
	}\n\
snippet tryf\n\
	try {\n\
		${3}\n\
	} catch(${1:Exception} ${2:e}) {\n\
	} finally {\n\
	}\n\
##\n\
## Find Methods\n\
snippet findall\n\
	List<${1:listName}> ${2:items} = ${1}.findAll();${3}\n\
snippet findbyid\n\
	${1:var} ${2:item} = ${1}.findById(${3});${4}\n\
##\n\
## Javadocs\n\
snippet /**\n\
	/**\n\
	 * ${1}\n\
	 */\n\
snippet @au\n\
	@author `system(\"grep \\`id -un\\` /etc/passwd | cut -d \\\":\\\" -f5 | cut -d \\\",\\\" -f1\")`\n\
snippet @br\n\
	@brief ${1:Description}\n\
snippet @fi\n\
	@file ${1:`Filename()`}.java\n\
snippet @pa\n\
	@param ${1:param}\n\
snippet @re\n\
	@return ${1:param}\n\
##\n\
## Logger Methods\n\
snippet debug\n\
	Logger.debug(${1:param});${2}\n\
snippet error\n\
	Logger.error(${1:param});${2}\n\
snippet info\n\
	Logger.info(${1:param});${2}\n\
snippet warn\n\
	Logger.warn(${1:param});${2}\n\
##\n\
## Loops\n\
snippet enfor\n\
	for (${1} : ${2}) ${3}\n\
snippet for\n\
	for (${1}; ${2}; ${3}) ${4}\n\
snippet wh\n\
	while (${1}) ${2}\n\
##\n\
## Main method\n\
snippet main\n\
	public static void main (String[] args) {\n\
		${1:/* code */}\n\
	}\n\
##\n\
## Print Methods\n\
snippet print\n\
	System.out.print(\"${1:Message}\");\n\
snippet printf\n\
	System.out.printf(\"${1:Message}\", ${2:args});\n\
snippet println\n\
	System.out.println(${1});\n\
##\n\
## Render Methods\n\
snippet ren\n\
	render(${1:param});${2}\n\
snippet rena\n\
	renderArgs.put(\"${1}\", ${2});${3}\n\
snippet renb\n\
	renderBinary(${1:param});${2}\n\
snippet renj\n\
	renderJSON(${1:param});${2}\n\
snippet renx\n\
	renderXml(${1:param});${2}\n\
##\n\
## Setter and Getter Methods\n\
snippet set\n\
	${1:public} void set${3:}(${2:String} ${4:}){\n\
		this.$4 = $4;\n\
	}\n\
snippet get\n\
	${1:public} ${2:String} get${3:}(){\n\
		return this.${4:};\n\
	}\n\
##\n\
## Terminate Methods or Loops\n\
snippet re\n\
	return\n\
snippet br\n\
	break;\n\
##\n\
## Test Methods\n\
snippet t\n\
	public void test${1:Name}() throws Exception {\n\
		${2}\n\
	}\n\
snippet test\n\
	@Test\n\
	public void test${1:Name}() throws Exception {\n\
		${2}\n\
	}\n\
##\n\
## Utils\n\
snippet Sc\n\
	Scanner\n\
##\n\
## Miscellaneous\n\
snippet action\n\
	public static void ${1:index}(${2:args}) { ${3} }\n\
snippet rnf\n\
	notFound(${1:param});${2}\n\
snippet rnfin\n\
	notFoundIfNull(${1:param});${2}\n\
snippet rr\n\
	redirect(${1:param});${2}\n\
snippet ru\n\
	unauthorized(${1:param});${2}\n\
snippet unless\n\
	(unless=${1:param});${2}\n\
";
exports.scope = "java";

});
