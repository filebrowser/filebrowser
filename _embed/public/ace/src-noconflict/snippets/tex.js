ace.define("ace/snippets/tex",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "#PREAMBLE\n\
#newcommand\n\
snippet nc\n\
	\\newcommand{\\${1:cmd}}[${2:opt}]{${3:realcmd}}${4}\n\
#usepackage\n\
snippet up\n\
	\\usepackage[${1:[options}]{${2:package}}\n\
#newunicodechar\n\
snippet nuc\n\
	\\newunicodechar{${1}}{${2:\\ensuremath}${3:tex-substitute}}}\n\
#DeclareMathOperator\n\
snippet dmo\n\
	\\DeclareMathOperator{${1}}{${2}}\n\
\n\
#DOCUMENT\n\
# \\begin{}...\\end{}\n\
snippet begin\n\
	\\begin{${1:env}}\n\
		${2}\n\
	\\end{$1}\n\
# Tabular\n\
snippet tab\n\
	\\begin{${1:tabular}}{${2:c}}\n\
	${3}\n\
	\\end{$1}\n\
snippet thm\n\
	\\begin[${1:author}]{${2:thm}}\n\
	${3}\n\
	\\end{$1}\n\
snippet center\n\
	\\begin{center}\n\
		${1}\n\
	\\end{center}\n\
# Align(ed)\n\
snippet ali\n\
	\\begin{align${1:ed}}\n\
		${2}\n\
	\\end{align$1}\n\
# Gather(ed)\n\
snippet gat\n\
	\\begin{gather${1:ed}}\n\
		${2}\n\
	\\end{gather$1}\n\
# Equation\n\
snippet eq\n\
	\\begin{equation}\n\
		${1}\n\
	\\end{equation}\n\
# Equation\n\
snippet eq*\n\
	\\begin{equation*}\n\
		${1}\n\
	\\end{equation*}\n\
# Unnumbered Equation\n\
snippet \\\n\
	\\[\n\
		${1}\n\
	\\]\n\
# Enumerate\n\
snippet enum\n\
	\\begin{enumerate}\n\
		\\item ${1}\n\
	\\end{enumerate}\n\
# Itemize\n\
snippet itemize\n\
	\\begin{itemize}\n\
		\\item ${1}\n\
	\\end{itemize}\n\
# Description\n\
snippet desc\n\
	\\begin{description}\n\
		\\item[${1}] ${2}\n\
	\\end{description}\n\
# Matrix\n\
snippet mat\n\
	\\begin{${1:p/b/v/V/B/small}matrix}\n\
		${2}\n\
	\\end{$1matrix}\n\
# Cases\n\
snippet cas\n\
	\\begin{cases}\n\
		${1:equation}, &\\text{ if }${2:case}\\\\\n\
		${3}\n\
	\\end{cases}\n\
# Split\n\
snippet spl\n\
	\\begin{split}\n\
		${1}\n\
	\\end{split}\n\
# Part\n\
snippet part\n\
	\\part{${1:part name}} % (fold)\n\
	\\label{prt:${2:$1}}\n\
	${3}\n\
	% part $2 (end)\n\
# Chapter\n\
snippet cha\n\
	\\chapter{${1:chapter name}}\n\
	\\label{cha:${2:$1}}\n\
	${3}\n\
# Section\n\
snippet sec\n\
	\\section{${1:section name}}\n\
	\\label{sec:${2:$1}}\n\
	${3}\n\
# Sub Section\n\
snippet sub\n\
	\\subsection{${1:subsection name}}\n\
	\\label{sub:${2:$1}}\n\
	${3}\n\
# Sub Sub Section\n\
snippet subs\n\
	\\subsubsection{${1:subsubsection name}}\n\
	\\label{ssub:${2:$1}}\n\
	${3}\n\
# Paragraph\n\
snippet par\n\
	\\paragraph{${1:paragraph name}}\n\
	\\label{par:${2:$1}}\n\
	${3}\n\
# Sub Paragraph\n\
snippet subp\n\
	\\subparagraph{${1:subparagraph name}}\n\
	\\label{subp:${2:$1}}\n\
	${3}\n\
#References\n\
snippet itd\n\
	\\item[${1:description}] ${2:item}\n\
snippet figure\n\
	${1:Figure}~\\ref{${2:fig:}}${3}\n\
snippet table\n\
	${1:Table}~\\ref{${2:tab:}}${3}\n\
snippet listing\n\
	${1:Listing}~\\ref{${2:list}}${3}\n\
snippet section\n\
	${1:Section}~\\ref{${2:sec:}}${3}\n\
snippet page\n\
	${1:page}~\\pageref{${2}}${3}\n\
snippet index\n\
	\\index{${1:index}}${2}\n\
#Citations\n\
snippet cite\n\
	\\cite[${1}]{${2}}${3}\n\
snippet fcite\n\
	\\footcite[${1}]{${2}}${3}\n\
#Formating text: italic, bold, underline, small capital, emphase ..\n\
snippet it\n\
	\\textit{${1:text}}\n\
snippet bf\n\
	\\textbf{${1:text}}\n\
snippet under\n\
	\\underline{${1:text}}\n\
snippet emp\n\
	\\emph{${1:text}}\n\
snippet sc\n\
	\\textsc{${1:text}}\n\
#Choosing font\n\
snippet sf\n\
	\\textsf{${1:text}}\n\
snippet rm\n\
	\\textrm{${1:text}}\n\
snippet tt\n\
	\\texttt{${1:text}}\n\
#misc\n\
snippet ft\n\
	\\footnote{${1:text}}\n\
snippet fig\n\
	\\begin{figure}\n\
	\\begin{center}\n\
	    \\includegraphics[scale=${1}]{Figures/${2}}\n\
	\\end{center}\n\
	\\caption{${3}}\n\
	\\label{fig:${4}}\n\
	\\end{figure}\n\
snippet tikz\n\
	\\begin{figure}\n\
	\\begin{center}\n\
	\\begin{tikzpicture}[scale=${1:1}]\n\
		${2}\n\
	\\end{tikzpicture}\n\
	\\end{center}\n\
	\\caption{${3}}\n\
	\\label{fig:${4}}\n\
	\\end{figure}\n\
#math\n\
snippet stackrel\n\
	\\stackrel{${1:above}}{${2:below}} ${3}\n\
snippet frac\n\
	\\frac{${1:num}}{${2:denom}}\n\
snippet sum\n\
	\\sum^{${1:n}}_{${2:i=1}}{${3}}";
exports.scope = "tex";

});
