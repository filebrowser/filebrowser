({
    optimize: "none",
    preserveLicenseComments: false,
    name: "node_modules/almond/almond",
    baseUrl: "../../",
    paths: {
        ace : "lib/ace",
        demo: "demo/kitchen-sink"        
    },
    packages: [
    ],
    include: [
        "ace/ace"
    ],
    exclude: [
    ],
    out: "./packed.js",
    useStrict: true,
    wrap: false
})<!DOCTYPE html>

<html>
<head>
    <title>Editor</title>
    <link rel="stylesheet" href="../kitchen-sink/styles.css" type="text/css" media="screen" charset="utf-8">    
</head>
<body>
    <div id="optionsPanel" style="position:absolute;height:100%;width:260px">
        <a href="http://c9.io" title="Cloud9 IDE | Your code anywhere, anytime">
            <img id="c9-logo" src="../kitchen-sink/logo.png" style="width: 172px;margin: -9px 30px -12px 51px;">
        </a>
    </div>
    <pre id="editor-container">
        <div style="color:black; padding: 10px">
            demo showing Ace usage with r.js:
            
            install r.js and almond
            and run `<code>r.js -o demo/r.js/build.js</code>`
            
            note that you also need ace/build/src to lazy load modes and themes
            require("ace/config").set("basePath", "../../build/src");
            require("ace/config").set("packaged", true);
        <div>
    </pre>

    <script src="./packed.js" data-ace-base="src" type="text/javascript" charset="utf-8"></script>  
    <script type="text/javascript" charset="utf-8">
        require("ace/config").set("basePath", "../../build/src")
        require("ace/config").set("packaged", true)
        var editor = require("ace/ace").edit("editor-container");
        editor.session.setMode("ace/mode/javascript");
        // editor.session.setValue("var editor = Ace!")
    </script>
</body>
</html>
