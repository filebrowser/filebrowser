define("ace/snippets/html",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# Some useful Unicode entities\n\
# Non-Breaking Space\n\
snippet nbs\n\
	&nbsp;\n\
# ←\n\
snippet left\n\
	&#x2190;\n\
# →\n\
snippet right\n\
	&#x2192;\n\
# ↑\n\
snippet up\n\
	&#x2191;\n\
# ↓\n\
snippet down\n\
	&#x2193;\n\
# ↩\n\
snippet return\n\
	&#x21A9;\n\
# ⇤\n\
snippet backtab\n\
	&#x21E4;\n\
# ⇥\n\
snippet tab\n\
	&#x21E5;\n\
# ⇧\n\
snippet shift\n\
	&#x21E7;\n\
# ⌃\n\
snippet ctrl\n\
	&#x2303;\n\
# ⌅\n\
snippet enter\n\
	&#x2305;\n\
# ⌘\n\
snippet cmd\n\
	&#x2318;\n\
# ⌥\n\
snippet option\n\
	&#x2325;\n\
# ⌦\n\
snippet delete\n\
	&#x2326;\n\
# ⌫\n\
snippet backspace\n\
	&#x232B;\n\
# ⎋\n\
snippet esc\n\
	&#x238B;\n\
# Generic Doctype\n\
snippet doctype HTML 4.01 Strict\n\
	<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\"\n\
	\"http://www.w3.org/TR/html4/strict.dtd\">\n\
snippet doctype HTML 4.01 Transitional\n\
	<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\"\n\
	\"http://www.w3.org/TR/html4/loose.dtd\">\n\
snippet doctype HTML 5\n\
	<!DOCTYPE HTML>\n\
snippet doctype XHTML 1.0 Frameset\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\"\n\
	\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">\n\
snippet doctype XHTML 1.0 Strict\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\"\n\
	\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">\n\
snippet doctype XHTML 1.0 Transitional\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\"\n\
	\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">\n\
snippet doctype XHTML 1.1\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.1//EN\"\n\
	\"http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd\">\n\
# HTML Doctype 4.01 Strict\n\
snippet docts\n\
	<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\"\n\
	\"http://www.w3.org/TR/html4/strict.dtd\">\n\
# HTML Doctype 4.01 Transitional\n\
snippet doct\n\
	<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\"\n\
	\"http://www.w3.org/TR/html4/loose.dtd\">\n\
# HTML Doctype 5\n\
snippet doct5\n\
	<!DOCTYPE HTML>\n\
# XHTML Doctype 1.0 Frameset\n\
snippet docxf\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Frameset//EN\"\n\
	\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd\">\n\
# XHTML Doctype 1.0 Strict\n\
snippet docxs\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\"\n\
	\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">\n\
# XHTML Doctype 1.0 Transitional\n\
snippet docxt\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\"\n\
	\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">\n\
# XHTML Doctype 1.1\n\
snippet docx\n\
	<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.1//EN\"\n\
	\"http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd\">\n\
# Attributes\n\
snippet attr\n\
	${1:attribute}=\"${2:property}\"\n\
snippet attr+\n\
	${1:attribute}=\"${2:property}\" attr+${3}\n\
snippet .\n\
	class=\"${1}\"${2}\n\
snippet #\n\
	id=\"${1}\"${2}\n\
snippet alt\n\
	alt=\"${1}\"${2}\n\
snippet charset\n\
	charset=\"${1:utf-8}\"${2}\n\
snippet data\n\
	data-${1}=\"${2:$1}\"${3}\n\
snippet for\n\
	for=\"${1}\"${2}\n\
snippet height\n\
	height=\"${1}\"${2}\n\
snippet href\n\
	href=\"${1:#}\"${2}\n\
snippet lang\n\
	lang=\"${1:en}\"${2}\n\
snippet media\n\
	media=\"${1}\"${2}\n\
snippet name\n\
	name=\"${1}\"${2}\n\
snippet rel\n\
	rel=\"${1}\"${2}\n\
snippet scope\n\
	scope=\"${1:row}\"${2}\n\
snippet src\n\
	src=\"${1}\"${2}\n\
snippet title=\n\
	title=\"${1}\"${2}\n\
snippet type\n\
	type=\"${1}\"${2}\n\
snippet value\n\
	value=\"${1}\"${2}\n\
snippet width\n\
	width=\"${1}\"${2}\n\
# Elements\n\
snippet a\n\
	<a href=\"${1:#}\">${2:$1}</a>\n\
snippet a.\n\
	<a class=\"${1}\" href=\"${2:#}\">${3:$1}</a>\n\
snippet a#\n\
	<a id=\"${1}\" href=\"${2:#}\">${3:$1}</a>\n\
snippet a:ext\n\
	<a href=\"http://${1:example.com}\">${2:$1}</a>\n\
snippet a:mail\n\
	<a href=\"mailto:${1:joe@example.com}?subject=${2:feedback}\">${3:email me}</a>\n\
snippet abbr\n\
	<abbr title=\"${1}\">${2}</abbr>\n\
snippet address\n\
	<address>\n\
		${1}\n\
	</address>\n\
snippet area\n\
	<area shape=\"${1:rect}\" coords=\"${2}\" href=\"${3}\" alt=\"${4}\" />\n\
snippet area+\n\
	<area shape=\"${1:rect}\" coords=\"${2}\" href=\"${3}\" alt=\"${4}\" />\n\
	area+${5}\n\
snippet area:c\n\
	<area shape=\"circle\" coords=\"${1}\" href=\"${2}\" alt=\"${3}\" />\n\
snippet area:d\n\
	<area shape=\"default\" coords=\"${1}\" href=\"${2}\" alt=\"${3}\" />\n\
snippet area:p\n\
	<area shape=\"poly\" coords=\"${1}\" href=\"${2}\" alt=\"${3}\" />\n\
snippet area:r\n\
	<area shape=\"rect\" coords=\"${1}\" href=\"${2}\" alt=\"${3}\" />\n\
snippet article\n\
	<article>\n\
		${1}\n\
	</article>\n\
snippet article.\n\
	<article class=\"${1}\">\n\
		${2}\n\
	</article>\n\
snippet article#\n\
	<article id=\"${1}\">\n\
		${2}\n\
	</article>\n\
snippet aside\n\
	<aside>\n\
		${1}\n\
	</aside>\n\
snippet aside.\n\
	<aside class=\"${1}\">\n\
		${2}\n\
	</aside>\n\
snippet aside#\n\
	<aside id=\"${1}\">\n\
		${2}\n\
	</aside>\n\
snippet audio\n\
	<audio src=\"${1}>${2}</audio>\n\
snippet b\n\
	<b>${1}</b>\n\
snippet base\n\
	<base href=\"${1}\" target=\"${2}\" />\n\
snippet bdi\n\
	<bdi>${1}</bdo>\n\
snippet bdo\n\
	<bdo dir=\"${1}\">${2}</bdo>\n\
snippet bdo:l\n\
	<bdo dir=\"ltr\">${1}</bdo>\n\
snippet bdo:r\n\
	<bdo dir=\"rtl\">${1}</bdo>\n\
snippet blockquote\n\
	<blockquote>\n\
		${1}\n\
	</blockquote>\n\
snippet body\n\
	<body>\n\
		${1}\n\
	</body>\n\
snippet br\n\
	<br />${1}\n\
snippet button\n\
	<button type=\"${1:submit}\">${2}</button>\n\
snippet button.\n\
	<button class=\"${1:button}\" type=\"${2:submit}\">${3}</button>\n\
snippet button#\n\
	<button id=\"${1}\" type=\"${2:submit}\">${3}</button>\n\
snippet button:s\n\
	<button type=\"submit\">${1}</button>\n\
snippet button:r\n\
	<button type=\"reset\">${1}</button>\n\
snippet canvas\n\
	<canvas>\n\
		${1}\n\
	</canvas>\n\
snippet caption\n\
	<caption>${1}</caption>\n\
snippet cite\n\
	<cite>${1}</cite>\n\
snippet code\n\
	<code>${1}</code>\n\
snippet col\n\
	<col />${1}\n\
snippet col+\n\
	<col />\n\
	col+${1}\n\
snippet colgroup\n\
	<colgroup>\n\
		${1}\n\
	</colgroup>\n\
snippet colgroup+\n\
	<colgroup>\n\
		<col />\n\
		col+${1}\n\
	</colgroup>\n\
snippet command\n\
	<command type=\"command\" label=\"${1}\" icon=\"${2}\" />\n\
snippet command:c\n\
	<command type=\"checkbox\" label=\"${1}\" icon=\"${2}\" />\n\
snippet command:r\n\
	<command type=\"radio\" radiogroup=\"${1}\" label=\"${2}\" icon=\"${3}\" />\n\
snippet datagrid\n\
	<datagrid>\n\
		${1}\n\
	</datagrid>\n\
snippet datalist\n\
	<datalist>\n\
		${1}\n\
	</datalist>\n\
snippet datatemplate\n\
	<datatemplate>\n\
		${1}\n\
	</datatemplate>\n\
snippet dd\n\
	<dd>${1}</dd>\n\
snippet dd.\n\
	<dd class=\"${1}\">${2}</dd>\n\
snippet dd#\n\
	<dd id=\"${1}\">${2}</dd>\n\
snippet del\n\
	<del>${1}</del>\n\
snippet details\n\
	<details>${1}</details>\n\
snippet dfn\n\
	<dfn>${1}</dfn>\n\
snippet dialog\n\
	<dialog>\n\
		${1}\n\
	</dialog>\n\
snippet div\n\
	<div>\n\
		${1}\n\
	</div>\n\
snippet div.\n\
	<div class=\"${1}\">\n\
		${2}\n\
	</div>\n\
snippet div#\n\
	<div id=\"${1}\">\n\
		${2}\n\
	</div>\n\
snippet dl\n\
	<dl>\n\
		${1}\n\
	</dl>\n\
snippet dl.\n\
	<dl class=\"${1}\">\n\
		${2}\n\
	</dl>\n\
snippet dl#\n\
	<dl id=\"${1}\">\n\
		${2}\n\
	</dl>\n\
snippet dl+\n\
	<dl>\n\
		<dt>${1}</dt>\n\
		<dd>${2}</dd>\n\
		dt+${3}\n\
	</dl>\n\
snippet dt\n\
	<dt>${1}</dt>\n\
snippet dt.\n\
	<dt class=\"${1}\">${2}</dt>\n\
snippet dt#\n\
	<dt id=\"${1}\">${2}</dt>\n\
snippet dt+\n\
	<dt>${1}</dt>\n\
	<dd>${2}</dd>\n\
	dt+${3}\n\
snippet em\n\
	<em>${1}</em>\n\
snippet embed\n\
	<embed src=${1} type=\"${2} />\n\
snippet fieldset\n\
	<fieldset>\n\
		${1}\n\
	</fieldset>\n\
snippet fieldset.\n\
	<fieldset class=\"${1}\">\n\
		${2}\n\
	</fieldset>\n\
snippet fieldset#\n\
	<fieldset id=\"${1}\">\n\
		${2}\n\
	</fieldset>\n\
snippet fieldset+\n\
	<fieldset>\n\
		<legend><span>${1}</span></legend>\n\
		${2}\n\
	</fieldset>\n\
	fieldset+${3}\n\
snippet figcaption\n\
	<figcaption>${1}</figcaption>\n\
snippet figure\n\
	<figure>${1}</figure>\n\
snippet footer\n\
	<footer>\n\
		${1}\n\
	</footer>\n\
snippet footer.\n\
	<footer class=\"${1}\">\n\
		${2}\n\
	</footer>\n\
snippet footer#\n\
	<footer id=\"${1}\">\n\
		${2}\n\
	</footer>\n\
snippet form\n\
	<form action=\"${1}\" method=\"${2:get}\" accept-charset=\"utf-8\">\n\
		${3}\n\
	</form>\n\
snippet form.\n\
	<form class=\"${1}\" action=\"${2}\" method=\"${3:get}\" accept-charset=\"utf-8\">\n\
		${4}\n\
	</form>\n\
snippet form#\n\
	<form id=\"${1}\" action=\"${2}\" method=\"${3:get}\" accept-charset=\"utf-8\">\n\
		${4}\n\
	</form>\n\
snippet h1\n\
	<h1>${1}</h1>\n\
snippet h1.\n\
	<h1 class=\"${1}\">${2}</h1>\n\
snippet h1#\n\
	<h1 id=\"${1}\">${2}</h1>\n\
snippet h2\n\
	<h2>${1}</h2>\n\
snippet h2.\n\
	<h2 class=\"${1}\">${2}</h2>\n\
snippet h2#\n\
	<h2 id=\"${1}\">${2}</h2>\n\
snippet h3\n\
	<h3>${1}</h3>\n\
snippet h3.\n\
	<h3 class=\"${1}\">${2}</h3>\n\
snippet h3#\n\
	<h3 id=\"${1}\">${2}</h3>\n\
snippet h4\n\
	<h4>${1}</h4>\n\
snippet h4.\n\
	<h4 class=\"${1}\">${2}</h4>\n\
snippet h4#\n\
	<h4 id=\"${1}\">${2}</h4>\n\
snippet h5\n\
	<h5>${1}</h5>\n\
snippet h5.\n\
	<h5 class=\"${1}\">${2}</h5>\n\
snippet h5#\n\
	<h5 id=\"${1}\">${2}</h5>\n\
snippet h6\n\
	<h6>${1}</h6>\n\
snippet h6.\n\
	<h6 class=\"${1}\">${2}</h6>\n\
snippet h6#\n\
	<h6 id=\"${1}\">${2}</h6>\n\
snippet head\n\
	<head>\n\
		<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\" />\n\
\n\
		<title>${1:`substitute(Filename('', 'Page Title'), '^.', '\\u&', '')`}</title>\n\
		${2}\n\
	</head>\n\
snippet header\n\
	<header>\n\
		${1}\n\
	</header>\n\
snippet header.\n\
	<header class=\"${1}\">\n\
		${2}\n\
	</header>\n\
snippet header#\n\
	<header id=\"${1}\">\n\
		${2}\n\
	</header>\n\
snippet hgroup\n\
	<hgroup>\n\
		${1}\n\
	</hgroup>\n\
snippet hgroup.\n\
	<hgroup class=\"${1}>\n\
		${2}\n\
	</hgroup>\n\
snippet hr\n\
	<hr />${1}\n\
snippet html\n\
	<html>\n\
	${1}\n\
	</html>\n\
snippet xhtml\n\
	<html xmlns=\"http://www.w3.org/1999/xhtml\">\n\
	${1}\n\
	</html>\n\
snippet html5\n\
	<!DOCTYPE html>\n\
	<html>\n\
		<head>\n\
			<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\" />\n\
			<title>${1:`substitute(Filename('', 'Page Title'), '^.', '\\u&', '')`}</title>\n\
			${2:meta}\n\
		</head>\n\
		<body>\n\
			${3:body}\n\
		</body>\n\
	</html>\n\
snippet i\n\
	<i>${1}</i>\n\
snippet iframe\n\
	<iframe src=\"${1}\" frameborder=\"0\"></iframe>${2}\n\
snippet iframe.\n\
	<iframe class=\"${1}\" src=\"${2}\" frameborder=\"0\"></iframe>${3}\n\
snippet iframe#\n\
	<iframe id=\"${1}\" src=\"${2}\" frameborder=\"0\"></iframe>${3}\n\
snippet img\n\
	<img src=\"${1}\" alt=\"${2}\" />${3}\n\
snippet img.\n\
	<img class=\"${1}\" src=\"${2}\" alt=\"${3}\" />${4}\n\
snippet img#\n\
	<img id=\"${1}\" src=\"${2}\" alt=\"${3}\" />${4}\n\
snippet input\n\
	<input type=\"${1:text/submit/hidden/button/image}\" name=\"${2}\" id=\"${3:$2}\" value=\"${4}\" />${5}\n\
snippet input.\n\
	<input class=\"${1}\" type=\"${2:text/submit/hidden/button/image}\" name=\"${3}\" id=\"${4:$3}\" value=\"${5}\" />${6}\n\
snippet input:text\n\
	<input type=\"text\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:submit\n\
	<input type=\"submit\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:hidden\n\
	<input type=\"hidden\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:button\n\
	<input type=\"button\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:image\n\
	<input type=\"image\" name=\"${1}\" id=\"${2:$1}\" src=\"${3}\" alt=\"${4}\" />${5}\n\
snippet input:checkbox\n\
	<input type=\"checkbox\" name=\"${1}\" id=\"${2:$1}\" />${3}\n\
snippet input:radio\n\
	<input type=\"radio\" name=\"${1}\" id=\"${2:$1}\" />${3}\n\
snippet input:color\n\
	<input type=\"color\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:date\n\
	<input type=\"date\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:datetime\n\
	<input type=\"datetime\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:datetime-local\n\
	<input type=\"datetime-local\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:email\n\
	<input type=\"email\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:file\n\
	<input type=\"file\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:month\n\
	<input type=\"month\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:number\n\
	<input type=\"number\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:password\n\
	<input type=\"password\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:range\n\
	<input type=\"range\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:reset\n\
	<input type=\"reset\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:search\n\
	<input type=\"search\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:time\n\
	<input type=\"time\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:url\n\
	<input type=\"url\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet input:week\n\
	<input type=\"week\" name=\"${1}\" id=\"${2:$1}\" value=\"${3}\" />${4}\n\
snippet ins\n\
	<ins>${1}</ins>\n\
snippet kbd\n\
	<kbd>${1}</kbd>\n\
snippet keygen\n\
	<keygen>${1}</keygen>\n\
snippet label\n\
	<label for=\"${2:$1}\">${1}</label>\n\
snippet label:i\n\
	<label for=\"${2:$1}\">${1}</label>\n\
	<input type=\"${3:text/submit/hidden/button}\" name=\"${4:$2}\" id=\"${5:$2}\" value=\"${6}\" />${7}\n\
snippet label:s\n\
	<label for=\"${2:$1}\">${1}</label>\n\
	<select name=\"${3:$2}\" id=\"${4:$2}\">\n\
		<option value=\"${5}\">${6:$5}</option>\n\
	</select>\n\
snippet legend\n\
	<legend>${1}</legend>\n\
snippet legend+\n\
	<legend><span>${1}</span></legend>\n\
snippet li\n\
	<li>${1}</li>\n\
snippet li.\n\
	<li class=\"${1}\">${2}</li>\n\
snippet li+\n\
	<li>${1}</li>\n\
	li+${2}\n\
snippet lia\n\
	<li><a href=\"${2:#}\">${1}</a></li>\n\
snippet lia+\n\
	<li><a href=\"${2:#}\">${1}</a></li>\n\
	lia+${3}\n\
snippet link\n\
	<link rel=\"${1}\" href=\"${2}\" title=\"${3}\" type=\"${4}\" />${5}\n\
snippet link:atom\n\
	<link rel=\"alternate\" href=\"${1:atom.xml}\" title=\"Atom\" type=\"application/atom+xml\" />${2}\n\
snippet link:css\n\
	<link rel=\"stylesheet\" href=\"${2:style.css}\" type=\"text/css\" media=\"${3:all}\" />${4}\n\
snippet link:favicon\n\
	<link rel=\"shortcut icon\" href=\"${1:favicon.ico}\" type=\"image/x-icon\" />${2}\n\
snippet link:rss\n\
	<link rel=\"alternate\" href=\"${1:rss.xml}\" title=\"RSS\" type=\"application/atom+xml\" />${2}\n\
snippet link:touch\n\
	<link rel=\"apple-touch-icon\" href=\"${1:favicon.png}\" />${2}\n\
snippet map\n\
	<map name=\"${1}\">\n\
		${2}\n\
	</map>\n\
snippet map.\n\
	<map class=\"${1}\" name=\"${2}\">\n\
		${3}\n\
	</map>\n\
snippet map#\n\
	<map name=\"${1}\" id=\"${2:$1}>\n\
		${3}\n\
	</map>\n\
snippet map+\n\
	<map name=\"${1}\">\n\
		<area shape=\"${2}\" coords=\"${3}\" href=\"${4}\" alt=\"${5}\" />${6}\n\
	</map>${7}\n\
snippet mark\n\
	<mark>${1}</mark>\n\
snippet menu\n\
	<menu>\n\
		${1}\n\
	</menu>\n\
snippet menu:c\n\
	<menu type=\"context\">\n\
		${1}\n\
	</menu>\n\
snippet menu:t\n\
	<menu type=\"toolbar\">\n\
		${1}\n\
	</menu>\n\
snippet meta\n\
	<meta http-equiv=\"${1}\" content=\"${2}\" />${3}\n\
snippet meta:compat\n\
	<meta http-equiv=\"X-UA-Compatible\" content=\"IE=${1:7,8,edge}\" />${3}\n\
snippet meta:refresh\n\
	<meta http-equiv=\"refresh\" content=\"text/html;charset=UTF-8\" />${3}\n\
snippet meta:utf\n\
	<meta http-equiv=\"content-type\" content=\"text/html;charset=UTF-8\" />${3}\n\
snippet meter\n\
	<meter>${1}</meter>\n\
snippet nav\n\
	<nav>\n\
		${1}\n\
	</nav>\n\
snippet nav.\n\
	<nav class=\"${1}\">\n\
		${2}\n\
	</nav>\n\
snippet nav#\n\
	<nav id=\"${1}\">\n\
		${2}\n\
	</nav>\n\
snippet noscript\n\
	<noscript>\n\
		${1}\n\
	</noscript>\n\
snippet object\n\
	<object data=\"${1}\" type=\"${2}\">\n\
		${3}\n\
	</object>${4}\n\
# Embed QT Movie\n\
snippet movie\n\
	<object width=\"$2\" height=\"$3\" classid=\"clsid:02BF25D5-8C17-4B23-BC80-D3488ABDDC6B\"\n\
	 codebase=\"http://www.apple.com/qtactivex/qtplugin.cab\">\n\
		<param name=\"src\" value=\"$1\" />\n\
		<param name=\"controller\" value=\"$4\" />\n\
		<param name=\"autoplay\" value=\"$5\" />\n\
		<embed src=\"${1:movie.mov}\"\n\
			width=\"${2:320}\" height=\"${3:240}\"\n\
			controller=\"${4:true}\" autoplay=\"${5:true}\"\n\
			scale=\"tofit\" cache=\"true\"\n\
			pluginspage=\"http://www.apple.com/quicktime/download/\" />\n\
	</object>${6}\n\
snippet ol\n\
	<ol>\n\
		${1}\n\
	</ol>\n\
snippet ol.\n\
	<ol class=\"${1}>\n\
		${2}\n\
	</ol>\n\
snippet ol#\n\
	<ol id=\"${1}>\n\
		${2}\n\
	</ol>\n\
snippet ol+\n\
	<ol>\n\
		<li>${1}</li>\n\
		li+${2}\n\
	</ol>\n\
snippet opt\n\
	<option value=\"${1}\">${2:$1}</option>\n\
snippet opt+\n\
	<option value=\"${1}\">${2:$1}</option>\n\
	opt+${3}\n\
snippet optt\n\
	<option>${1}</option>\n\
snippet optgroup\n\
	<optgroup>\n\
		<option value=\"${1}\">${2:$1}</option>\n\
		opt+${3}\n\
	</optgroup>\n\
snippet output\n\
	<output>${1}</output>\n\
snippet p\n\
	<p>${1}</p>\n\
snippet param\n\
	<param name=\"${1}\" value=\"${2}\" />${3}\n\
snippet pre\n\
	<pre>\n\
		${1}\n\
	</pre>\n\
snippet progress\n\
	<progress>${1}</progress>\n\
snippet q\n\
	<q>${1}</q>\n\
snippet rp\n\
	<rp>${1}</rp>\n\
snippet rt\n\
	<rt>${1}</rt>\n\
snippet ruby\n\
	<ruby>\n\
		<rp><rt>${1}</rt></rp>\n\
	</ruby>\n\
snippet s\n\
	<s>${1}</s>\n\
snippet samp\n\
	<samp>\n\
		${1}\n\
	</samp>\n\
snippet script\n\
	<script type=\"text/javascript\" charset=\"utf-8\">\n\
		${1}\n\
	</script>\n\
snippet scriptsrc\n\
	<script src=\"${1}.js\" type=\"text/javascript\" charset=\"utf-8\"></script>\n\
snippet section\n\
	<section>\n\
		${1}\n\
	</section>\n\
snippet section.\n\
	<section class=\"${1}\">\n\
		${2}\n\
	</section>\n\
snippet section#\n\
	<section id=\"${1}\">\n\
		${2}\n\
	</section>\n\
snippet select\n\
	<select name=\"${1}\" id=\"${2:$1}\">\n\
		${3}\n\
	</select>\n\
snippet select.\n\
	<select name=\"${1}\" id=\"${2:$1}\" class=\"${3}>\n\
		${4}\n\
	</select>\n\
snippet select+\n\
	<select name=\"${1}\" id=\"${2:$1}\">\n\
		<option value=\"${3}\">${4:$3}</option>\n\
		opt+${5}\n\
	</select>\n\
snippet small\n\
	<small>${1}</small>\n\
snippet source\n\
	<source src=\"${1}\" type=\"${2}\" media=\"${3}\" />\n\
snippet span\n\
	<span>${1}</span>\n\
snippet strong\n\
	<strong>${1}</strong>\n\
snippet style\n\
	<style type=\"text/css\" media=\"${1:all}\">\n\
		${2}\n\
	</style>\n\
snippet sub\n\
	<sub>${1}</sub>\n\
snippet summary\n\
	<summary>\n\
		${1}\n\
	</summary>\n\
snippet sup\n\
	<sup>${1}</sup>\n\
snippet table\n\
	<table border=\"${1:0}\">\n\
		${2}\n\
	</table>\n\
snippet table.\n\
	<table class=\"${1}\" border=\"${2:0}\">\n\
		${3}\n\
	</table>\n\
snippet table#\n\
	<table id=\"${1}\" border=\"${2:0}\">\n\
		${3}\n\
	</table>\n\
snippet tbody\n\
	<tbody>\n\
		${1}\n\
	</tbody>\n\
snippet td\n\
	<td>${1}</td>\n\
snippet td.\n\
	<td class=\"${1}\">${2}</td>\n\
snippet td#\n\
	<td id=\"${1}\">${2}</td>\n\
snippet td+\n\
	<td>${1}</td>\n\
	td+${2}\n\
snippet textarea\n\
	<textarea name=\"${1}\" id=${2:$1} rows=\"${3:8}\" cols=\"${4:40}\">${5}</textarea>${6}\n\
snippet tfoot\n\
	<tfoot>\n\
		${1}\n\
	</tfoot>\n\
snippet th\n\
	<th>${1}</th>\n\
snippet th.\n\
	<th class=\"${1}\">${2}</th>\n\
snippet th#\n\
	<th id=\"${1}\">${2}</th>\n\
snippet th+\n\
	<th>${1}</th>\n\
	th+${2}\n\
snippet thead\n\
	<thead>\n\
		${1}\n\
	</thead>\n\
snippet time\n\
	<time datetime=\"${1}\" pubdate=\"${2:$1}>${3:$1}</time>\n\
snippet title\n\
	<title>${1:`substitute(Filename('', 'Page Title'), '^.', '\\u&', '')`}</title>\n\
snippet tr\n\
	<tr>\n\
		${1}\n\
	</tr>\n\
snippet tr+\n\
	<tr>\n\
		<td>${1}</td>\n\
		td+${2}\n\
	</tr>\n\
snippet track\n\
	<track src=\"${1}\" srclang=\"${2}\" label=\"${3}\" default=\"${4:default}>${5}</track>${6}\n\
snippet ul\n\
	<ul>\n\
		${1}\n\
	</ul>\n\
snippet ul.\n\
	<ul class=\"${1}\">\n\
		${2}\n\
	</ul>\n\
snippet ul#\n\
	<ul id=\"${1}\">\n\
		${2}\n\
	</ul>\n\
snippet ul+\n\
	<ul>\n\
		<li>${1}</li>\n\
		li+${2}\n\
	</ul>\n\
snippet var\n\
	<var>${1}</var>\n\
snippet video\n\
	<video src=\"${1} height=\"${2}\" width=\"${3}\" preload=\"${5:none}\" autoplay=\"${6:autoplay}>${7}</video>${8}\n\
snippet wbr\n\
	<wbr />${1}\n\
";
exports.scope = "html";

});
