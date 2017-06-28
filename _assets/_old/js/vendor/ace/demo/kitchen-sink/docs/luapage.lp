<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
   "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html>
<% --[[--
    index.lp from the Kepler Project's LuaDoc HTML doclet.
    http://keplerproject.github.com/luadoc/
--]] %>
<head>
    <title>Reference</title>
    <link rel="stylesheet" href="<%=luadoc.doclet.html.link("luadoc.css")%>" type="text/css" />
	<!--meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/-->
</head>

<body>
<div id="container">

<div id="product">
	<div id="product_logo"></div>
	<div id="product_name"><big><b></b></big></div>
	<div id="product_description"></div>
</div> <!-- id="product" -->

<div id="main">

<div id="navigation">
<%=luadoc.doclet.html.include("menu.lp", { doc=doc })%>

</div> <!-- id="navigation" -->

<div id="content">


<%if not options.nomodules and #doc.modules > 0 then%>
<h2>Modules</h2>
<table class="module_list">
<!--<tr><td colspan="2">Modules</td></tr>-->
<%for _, modulename in ipairs(doc.modules) do%>
	<tr>
		<td class="name"><a href="<%=luadoc.doclet.html.module_link(modulename, doc)%>"><%=modulename%></a></td>
		<td class="summary"><%=doc.modules[modulename].summary%></td>
	</tr>
<%end%>
</table>
<%end%>



<%if not options.nofiles and #doc.files > 0 then%>
<h2>Files</h2>
<table class="file_list">
<!--<tr><td colspan="2">Files</td></tr>-->
<%for _, filepath in ipairs(doc.files) do%>
	<tr>
		<td class="name"><a href="<%=luadoc.doclet.html.file_link(filepath)%>"><%=filepath%></a></td>
		<td class="summary"></td>
	</tr>
<%end%>
</table>
<%end%>

</div> <!-- id="content" -->

</div> <!-- id="main" -->

<div id="about">
	<p><a href="http://validator.w3.org/check?uri=referer"><img src="http://www.w3.org/Icons/valid-xhtml10" alt="Valid XHTML 1.0!" height="31" width="88" /></a></p>
</div> <!-- id="about" -->

</div> <!-- id="container" -->	
</body>
</html>
