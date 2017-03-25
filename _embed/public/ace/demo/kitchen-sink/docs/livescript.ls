# Defines an editing mode for [Ace](http://ace.ajax.org).
#
# Open [test/ace.html](../test/ace.html) to test.

require, exports, module <-! define \ace/mode/ls

identifier = /(?![\d\s])[$\w\xAA-\uFFDC](?:(?!\s)[$\w\xAA-\uFFDC]|-[A-Za-z])*/$

exports.Mode = class LiveScriptMode extends require(\ace/mode/text)Mode
  ->
    @$tokenizer =
      new (require \ace/tokenizer)Tokenizer LiveScriptMode.Rules
    if require \ace/mode/matching_brace_outdent
      @$outdent = new that.MatchingBraceOutdent

  indenter = // (?
    : [({[=:]
    | [-~]>
    | \b (?: e(?:lse|xport) | d(?:o|efault) | t(?:ry|hen) | finally |
             import (?:\s* all)? | const | var |
             let | new | catch (?:\s* #identifier)? )
  ) \s* $ //

  getNextLineIndent: (state, line, tab) ->
    indent   = @$getIndent line
    {tokens} = @$tokenizer.getLineTokens line, state
    unless tokens.length and tokens[*-1]type is \comment
      indent += tab if state is \start and indenter.test line
    indent

  toggleCommentLines: (state, doc, startRow, endRow) ->
    comment = /^(\s*)#/; range = new (require \ace/range)Range 0 0 0 0
    for i from startRow to endRow
      if out = comment.test line = doc.getLine i
      then line.=replace comment, \$1
      else line.=replace /^\s*/   \$&#
      range.end.row = range.start.row = i
      range.end.column = line.length + 1
      doc.replace range, line
    1 - out * 2

  checkOutdent: (state, line, input) -> @$outdent?checkOutdent line, input

  autoOutdent: (state, doc, row) -> @$outdent?autoOutdent doc, row

### Highlight Rules

keywordend = /(?![$\w]|-[A-Za-z]|\s*:(?![:=]))/$
stringfill = token: \string, regex: '.+'

LiveScriptMode.Rules =
  start:
    * token: \keyword
      regex: //(?
        :t(?:h(?:is|row|en)|ry|ypeof!?)
        |c(?:on(?:tinue|st)|a(?:se|tch)|lass)
        |i(?:n(?:stanceof)?|mp(?:ort(?:\s+all)?|lements)|[fs])
        |d(?:e(?:fault|lete|bugger)|o)
        |f(?:or(?:\s+own)?|inally|unction)
        |s(?:uper|witch)
        |e(?:lse|x(?:tends|port)|val)
        |a(?:nd|rguments)
        |n(?:ew|ot)
        |un(?:less|til)
        |w(?:hile|ith)
        |o[fr]|return|break|let|var|loop
      )//$ + keywordend

    * token: \constant.language
      regex: '(?:true|false|yes|no|on|off|null|void|undefined)' + keywordend

    * token: \invalid.illegal
      regex: '(?
        :p(?:ackage|r(?:ivate|otected)|ublic)
        |i(?:mplements|nterface)
        |enum|static|yield
      )' + keywordend

    * token: \language.support.class
      regex: '(?
        :R(?:e(?:gExp|ferenceError)|angeError)
        |S(?:tring|yntaxError)
        |E(?:rror|valError)
        |Array|Boolean|Date|Function|Number|Object|TypeError|URIError
      )' + keywordend

    * token: \language.support.function
      regex: '(?
        :is(?:NaN|Finite)
        |parse(?:Int|Float)
        |Math|JSON
        |(?:en|de)codeURI(?:Component)?
      )' + keywordend

    * token: \variable.language
      regex: '(?:t(?:hat|il|o)|f(?:rom|allthrough)|it|by|e)' + keywordend

    * token: \identifier
      regex: identifier + /\s*:(?![:=])/$

    * token: \variable
      regex: identifier

    * token: \keyword.operator
      regex: /(?:\.{3}|\s+\?)/$

    * token: \keyword.variable
      regex: /(?:@+|::|\.\.)/$
      next : \key

    * token: \keyword.operator
      regex: /\.\s*/$
      next : \key

    * token: \string
      regex: /\\\S[^\s,;)}\]]*/$

    * token: \string.doc
      regex: \'''
      next : \qdoc

    * token: \string.doc
      regex: \"""
      next : \qqdoc

    * token: \string
      regex: \'
      next : \qstring

    * token: \string
      regex: \"
      next : \qqstring

    * token: \string
      regex: \`
      next : \js

    * token: \string
      regex: '<\\['
      next : \words

    * token: \string.regex
      regex: \//
      next : \heregex

    * token: \comment.doc
      regex: '/\\*'
      next : \comment

    * token: \comment
      regex: '#.*'

    * token: \string.regex
      regex: //
        /(?: [^ [ / \n \\ ]*
          (?: (?: \\.
                | \[ [^\]\n\\]* (?:\\.[^\]\n\\]*)* \]
              ) [^ [ / \n \\ ]*
          )*
        )/ [gimy$]{0,4}
      //$
      next : \key

    * token: \constant.numeric
      regex: '(?:0x[\\da-fA-F][\\da-fA-F_]*
                |(?:[2-9]|[12]\\d|3[0-6])r[\\da-zA-Z][\\da-zA-Z_]*
                |(?:\\d[\\d_]*(?:\\.\\d[\\d_]*)?|\\.\\d[\\d_]*)
                 (?:e[+-]?\\d[\\d_]*)?[\\w$]*)'

    * token: \lparen
      regex: '[({[]'

    * token: \rparen
      regex: '[)}\\]]'
      next : \key

    * token: \keyword.operator
      regex: \\\S+

    * token: \text
      regex: \\\s+

  heregex:
    * token: \string.regex
      regex: '.*?//[gimy$?]{0,4}'
      next : \start
    * token: \string.regex
      regex: '\\s*#{'
    * token: \comment.regex
      regex: '\\s+(?:#.*)?'
    * token: \string.regex
      regex: '\\S+'

  key:
    * token: \keyword.operator
      regex: '[.?@!]+'
    * token: \identifier
      regex: identifier
      next : \start
    * token: \text
      regex: '.'
      next : \start

  comment:
    * token: \comment.doc
      regex: '.*?\\*/'
      next : \start
    * token: \comment.doc
      regex: '.+'

  qdoc:
    token: \string
    regex: ".*?'''"
    next : \key
    stringfill

  qqdoc:
    token: \string
    regex: '.*?"""'
    next : \key
    stringfill

  qstring:
    token: \string
    regex: /[^\\']*(?:\\.[^\\']*)*'/$
    next : \key
    stringfill

  qqstring:
    token: \string
    regex: /[^\\"]*(?:\\.[^\\"]*)*"/$
    next : \key
    stringfill

  js:
    token: \string
    regex: /[^\\`]*(?:\\.[^\\`]*)*`/$
    next : \key
    stringfill

  words:
    token: \string
    regex: '.*?\\]>'
    next : \key
    stringfill
