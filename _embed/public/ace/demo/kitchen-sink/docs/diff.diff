diff --git a/lib/ace/edit_session.js b/lib/ace/edit_session.js
index 23fc3fc..ed3b273 100644
--- a/lib/ace/edit_session.js
+++ b/lib/ace/edit_session.js
@@ -51,6 +51,7 @@ var TextMode = require("./mode/text").Mode;
 var Range = require("./range").Range;
 var Document = require("./document").Document;
 var BackgroundTokenizer = require("./background_tokenizer").BackgroundTokenizer;
+var SearchHighlight = require("./search_highlight").SearchHighlight;
 
 /**
  * class EditSession
@@ -307,6 +308,13 @@ var EditSession = function(text, mode) {
         return token;
     };
 
+    this.highlight = function(re) {
+        if (!this.$searchHighlight) {
+            var highlight = new SearchHighlight(null, "ace_selected-word", "text");
+            this.$searchHighlight = this.addDynamicMarker(highlight);
+        }
+        this.$searchHighlight.setRegexp(re);
+    }
     /**
     * EditSession.setUndoManager(undoManager)
     * - undoManager (UndoManager): The new undo manager
@@ -556,7 +564,8 @@ var EditSession = function(text, mode) {
             type : type || "line",
             renderer: typeof type == "function" ? type : null,
             clazz : clazz,
-            inFront: !!inFront
+            inFront: !!inFront,
+            id: id
         }
 
         if (inFront) {
diff --git a/lib/ace/editor.js b/lib/ace/editor.js
index 834e603..b27ec73 100644
--- a/lib/ace/editor.js
+++ b/lib/ace/editor.js
@@ -494,7 +494,7 @@ var Editor = function(renderer, session) {
      * Emitted when a selection has changed.
      **/
     this.onSelectionChange = function(e) {
-        var session = this.getSession();
+        var session = this.session;
 
         if (session.$selectionMarker) {
             session.removeMarker(session.$selectionMarker);
@@ -509,12 +509,40 @@ var Editor = function(renderer, session) {
             this.$updateHighlightActiveLine();
         }
 
-        var self = this;
-        if (this.$highlightSelectedWord && !this.$wordHighlightTimer)
-            this.$wordHighlightTimer = setTimeout(function() {
-                self.session.$mode.highlightSelection(self);
-                self.$wordHighlightTimer = null;
-            }, 30, this);
+        var re = this.$highlightSelectedWord && this.$getSelectionHighLightRegexp()
     };
diff --git a/lib/ace/search_highlight.js b/lib/ace/search_highlight.js
new file mode 100644
index 0000000..b2df779
--- /dev/null
+++ b/lib/ace/search_highlight.js
@@ -0,0 +1,3 @@
+new
+empty file