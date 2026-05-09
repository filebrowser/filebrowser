// frontend/src/ace-gcode.js
// Custom G-code syntax mode for ACE editor.
//
// Token names map onto standard Ace categories (keyword, comment,
// variable.parameter, support.function, ...) so the *active theme*
// paints them — picking Ambiance vs Monokai vs Solarized actually
// changes the gcode colors instead of being overridden.
// Multi-segment token names also produce per-axis CSS hooks
// (e.g. "variable.parameter.x" → ".ace_variable.ace_parameter.ace_x")
// for the few accent overrides we keep in Editor.vue.

ace.define(
  "ace/mode/gcode_highlight_rules",
  [
    "require",
    "exports",
    "module",
    "ace/lib/oop",
    "ace/mode/text_highlight_rules",
  ],
  function (require, exports) {
    "use strict";

    var oop = require("ace/lib/oop");
    var TextHighlightRules =
      require("ace/mode/text_highlight_rules").TextHighlightRules;

    var GcodeHighlightRules = function () {
      this.$rules = {
        start: [
          // ── Comments ────────────────────────────────────────────────────────
          { token: "comment.gcode", regex: "\\(.*?\\)" },
          { token: "comment.gcode", regex: ";.*$" },

          // ── Block numbers (N100, N0010) ─────────────────────────────────────
          // Plain "constant" so the theme paints it; we dim it with opacity
          // in CSS to keep the line numbers from competing with the codes.
          { token: "constant.other.block.gcode", regex: "\\bN[0-9]+\\b" },

          // ── Program markers — % (tape start/end) and O1234 (program number).
          { token: "keyword.control.marker.gcode", regex: "%|\\bO[0-9]+\\b" },

          // ── G-words: G0, G01, G28.1, etc.
          { token: "keyword.gword.gcode", regex: "\\bG[0-9]+(?:\\.[0-9]+)?\\b" },

          // ── M-codes: M3, M06, M30, etc.
          { token: "keyword.other.mcode.gcode", regex: "\\bM[0-9]+(?:\\.[0-9]+)?\\b" },

          // ── Axis / arc params ────────────────────────────────────────────────
          // X / I / A — first axis class
          {
            token: "variable.parameter.x.gcode",
            regex: "\\b[AIX][+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },
          // Y / J — second axis class
          {
            token: "variable.parameter.y.gcode",
            regex: "\\b[JY][+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },
          // Z / K / B — third axis class
          {
            token: "variable.parameter.z.gcode",
            regex: "\\b[BKZ][+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── Feed / speed / tool / offset codes ──────────────────────────────
          // F, S, H, D, T, HCC — covers feed rate, spindle speed, tool/offset calls
          {
            token: "support.function.feedspeed.gcode",
            regex: "\\b(?:HCC|Hcc|hcc|[FSDHT])[+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── Subprogram / dwell P values
          {
            token: "entity.name.function.subprog.gcode",
            regex: "\\bP[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── R parameter (radius / retract) — geometry-like, group with X/I.
          {
            token: "variable.parameter.x.gcode",
            regex: "\\bR[+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── Fallback: bare numbers ───────────────────────────────────────────
          { token: "constant.numeric", regex: "[+-]?[0-9]+(?:\\.[0-9]+)?\\b" },
        ],
      };

      this.normalizeRules();
    };

    oop.inherits(GcodeHighlightRules, TextHighlightRules);
    exports.GcodeHighlightRules = GcodeHighlightRules;
  }
);

ace.define(
  "ace/mode/gcode",
  [
    "require",
    "exports",
    "module",
    "ace/lib/oop",
    "ace/mode/text",
    "ace/mode/gcode_highlight_rules",
  ],
  function (require, exports) {
    "use strict";

    var oop = require("ace/lib/oop");
    var TextMode = require("ace/mode/text").Mode;
    var GcodeHighlightRules =
      require("ace/mode/gcode_highlight_rules").GcodeHighlightRules;

    var Mode = function () {
      this.HighlightRules = GcodeHighlightRules;
      this.lineCommentStart = ";";
    };
    oop.inherits(Mode, TextMode);

    (function () {
      this.$id = "ace/mode/gcode";
    }).call(Mode.prototype);

    exports.Mode = Mode;
  }
);
