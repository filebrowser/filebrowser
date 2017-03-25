if (typeof ace == "undefined" && typeof require == "undefined") {
    document.body.innerHTML = "<p style='padding: 20px 50px;'>couldn't find ace.js file, <br>"
        + "to build it run <code>node Makefile.dryice.js full<code>"
} else if (typeof ace == "undefined" && typeof require != "undefined") {
    require(["ace/ace"], setValue)
} else {
    require = ace.require;
    setValue()
}

function setValue() {
    require("ace/lib/net").get(document.baseURI, function(t){
        var el = document.getElementById("editor");
        el.env.editor.setValue(t, 1);
    })
}