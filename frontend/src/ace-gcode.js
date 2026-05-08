// frontend/src/ace-gcode.js
// Custom G-code syntax mode for ACE editor.
// Tokens are intentionally neutral for N-codes so they inherit
// the active theme's foreground color — works in both light and dark mode.

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
          // Parenthetical: (this is a comment)
          {
            token: "gcode.comment",
            regex: "\\(.*?\\)",
          },
          // Semicolon: ; this is a comment
          {
            token: "gcode.comment",
            regex: ";.*$",
          },

          // ── Block numbers ────────────────────────────────────────────────────
          // N100, N0010 — uses "gcode.block" token; CSS sets color: inherit
          // so it renders correctly in both light and dark Ace themes.
          {
            token: "gcode.block",
            regex: "\\bN[0-9]+\\b",
          },

          // ── Program markers ──────────────────────────────────────────────────
          // % (tape start/end) and O1234 (program number)
          {
            token: "gcode.marker",
            regex: "%|\\bO[0-9]+\\b",
          },

          // ── G-words (purple) ─────────────────────────────────────────────────
          // G0, G01, G28.1, etc.
          {
            token: "gcode.gword",
            regex: "\\bG[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── M-codes (yellow) ─────────────────────────────────────────────────
          // M3, M06, M30, etc.
          {
            token: "gcode.mcode",
            regex: "\\bM[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── Axis / arc params ────────────────────────────────────────────────
          // X / I / A  →  orange
          // NOTE: ACE does NOT uppercase input; use explicit uppercase only.
          //       lowercase i/j removed — they never appear in standard NC files.
          {
            token: "gcode.xparam",
            regex: "\\b[AIX][+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // Y / J  →  teal
          {
            token: "gcode.yparam",
            regex: "\\b[JY][+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // Z / K / B  →  blue
          {
            token: "gcode.zparam",
            regex: "\\b[BKZ][+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── Feed / speed / tool / offset codes (teal) ───────────────────────
          // F, S, H, D, T, HCC — covers feed rate, spindle speed, tool/offset calls
          {
            token: "gcode.feedspeed",
            regex: "\\b(?:HCC|Hcc|hcc|[FSDHT])[+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── Subprogram / dwell P values (light blue) ────────────────────────
          {
            token: "gcode.subprog",
            regex: "\\bP[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── R parameter (radius / retract) ──────────────────────────────────
          // Kept separate so it can be styled distinctly if desired.
          // Currently falls through to constant.numeric — add a token if needed.
          {
            token: "gcode.xparam", // reuse orange — R is geometry like X/I
            regex: "\\bR[+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },

          // ── Fallback: bare numbers ───────────────────────────────────────────
          // Any remaining numeric literal — styled with opacity in CSS so it
          // recedes behind named tokens without hardcoding a color.
          {
            token: "constant.numeric",
            regex: "[+-]?[0-9]+(?:\\.[0-9]+)?\\b",
          },
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
