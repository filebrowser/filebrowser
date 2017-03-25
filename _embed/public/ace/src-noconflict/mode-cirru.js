ace.define("ace/mode/cirru_highlight_rules",["require","exports","module","ace/lib/oop","ace/mode/text_highlight_rules"], function(require, exports, module) {
"use strict";

var oop = require("../lib/oop");
var TextHighlightRules = require("./text_highlight_rules").TextHighlightRules;
var CirruHighlightRules = function() {
    this.$rules = {
        start: [{
            token: 'constant.numeric',
            regex: /[\d\.]+/
        }, {
            token: 'comment.line.double-dash',
            regex: /--/,
            next: 'comment'
        }, {
            token: 'storage.modifier',
            regex: /\(/
        }, {
            token: 'storage.modifier',
            regex: /,/,
            next: 'line'
        }, {
            token: 'support.function',
            regex: /[^\(\)"\s]+/,
            next: 'line'
        }, {
            token: 'string.quoted.double',
            regex: /"/,
            next: 'string'
        }, {
            token: 'storage.modifier',
            regex: /\)/
        }],
        comment: [{
            token: 'comment.line.double-dash',
            regex: / +[^\n]+/,
            next: 'start'
        }],
        string: [{
            token: 'string.quoted.double',
            regex: /"/,
            next: 'line'
        }, {
            token: 'constant.character.escape',
            regex: /\\/,
            next: 'escape'
        }, {
            token: 'string.quoted.double',
            regex: /[^\\"]+/
        }],
        escape: [{
            token: 'constant.character.escape',
            regex: /./,
            next: 'string'
        }],
        line: [{
            token: 'constant.numeric',
            regex: /[\d\.]+/
        }, {
            token: 'markup.raw',
            regex: /^\s*/,
            next: 'start'
        }, {
            token: 'storage.modifier',
            regex: /\$/,
            next: 'start'
        }, {
            token: 'variable.parameter',
            regex: /[^\(\)"\s]+/
        }, {
            token: 'storage.modifier',
            regex: /\(/,
            next: 'start'
        }, {
            token: 'storage.modifier',
            regex: /\)/
        }, {
            token: 'markup.raw',
            regex: /^ */,
            next: 'start'
        }, {
            token: 'string.quoted.double',
            regex: /"/,
            next: 'string'
        }]
    }

};

oop.inherits(CirruHighlightRules, TextHighlightRules);

exports.CirruHighlightRules = CirruHighlightRules;
});

ace.define("ace/mode/folding/coffee",["require","exports","module","ace/lib/oop","ace/mode/folding/fold_mode","ace/range"], function(require, exports, module) {
"use strict";

var oop = require("../../lib/oop");
var BaseFoldMode = require("./fold_mode").FoldMode;
var Range = require("../../range").Range;

var FoldMode = exports.FoldMode = function() {};
oop.inherits(FoldMode, BaseFoldMode);

(function() {

    this.getFoldWidgetRange = function(session, foldStyle, row) {
        var range = this.indentationBlock(session, row);
        if (range)
            return range;

        var re = /\S/;
        var line = session.getLine(row);
        var startLevel = line.search(re);
        if (startLevel == -1 || line[startLevel] != "#")
            return;

        var startColumn = line.length;
        var maxRow = session.getLength();
        var startRow = row;
        var endRow = row;

        while (++row < maxRow) {
            line = session.getLine(row);
            var level = line.search(re);

            if (level == -1)
                continue;

            if (line[level] != "#")
                break;

            endRow = row;
        }

        if (endRow > startRow) {
            var endColumn = session.getLine(endRow).length;
            return new Range(startRow, startColumn, endRow, endColumn);
        }
    };
    this.getFoldWidget = function(session, foldStyle, row) {
        var line = session.getLine(row);
        var indent = line.search(/\S/);
        var next = session.getLine(row + 1);
        var prev = session.getLine(row - 1);
        var prevIndent = prev.search(/\S/);
        var nextIndent = next.search(/\S/);

        if (indent == -1) {
            session.foldWidgets[row - 1] = prevIndent!= -1 && prevIndent < nextIndent ? "start" : "";
            return "";
        }
        if (prevIndent == -1) {
            if (indent == nextIndent && line[indent] == "#" && next[indent] == "#") {
                session.foldWidgets[row - 1] = "";
                session.foldWidgets[row + 1] = "";
                return "start";
            }
        } else if (prevIndent == indent && line[indent] == "#" && prev[indent] == "#") {
            if (session.getLine(row - 2).search(/\S/) == -1) {
                session.foldWidgets[row - 1] = "start";
                session.foldWidgets[row + 1] = "";
                return "";
            }
        }

        if (prevIndent!= -1 && prevIndent < indent)
            session.foldWidgets[row - 1] = "start";
        else
            session.foldWidgets[row - 1] = "";

        if (indent < nextIndent)
            return "start";
        else
            return "";
    };

}).call(FoldMode.prototype);

});

ace.define("ace/mode/cirru",["require","exports","module","ace/lib/oop","ace/mode/text","ace/mode/cirru_highlight_rules","ace/mode/folding/coffee"], function(require, exports, module) {
"use strict";

var oop = require("../lib/oop");
var TextMode = require("./text").Mode;
var CirruHighlightRules = require("./cirru_highlight_rules").CirruHighlightRules;
var CoffeeFoldMode = require("./folding/coffee").FoldMode;

var Mode = function() {
    this.HighlightRules = CirruHighlightRules;
    this.foldingRules = new CoffeeFoldMode();
    this.$behaviour = this.$defaultBehaviour;
};
oop.inherits(Mode, TextMode);

(function() {
    this.lineCommentStart = "--";
    this.$id = "ace/mode/cirru";
}).call(Mode.prototype);

exports.Mode = Mode;
});
