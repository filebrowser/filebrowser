<svg
  width="800" height="600"
  xmlns="http://www.w3.org/2000/svg"
  onload="StartAnimation(evt)">

  <title>Test Tube Progress Bar</title>
  <desc>Created for the Web Directions SVG competition</desc>

  <script type="text/ecmascript"><![CDATA[
    var timevalue = 0;
    var timer_increment = 1;
    var max_time = 100;
    var hickory;
    var dickory;
    var dock;
    var i;

    function StartAnimation(evt) {
        hickory  = evt.target.ownerDocument.getElementById("hickory");
        dickory = evt.target.ownerDocument.getElementById("dickory");
        dock = evt.target.ownerDocument.getElementById("dock");

        ShowAndGrowElement();
    }
    function ShowAndGrowElement() {
        timevalue = timevalue + timer_increment;
        if (timevalue > max_time)
            return;
        // Scale the text string gradually until it is 20 times larger
        scalefactor = (timevalue * 650) / max_time;

        if (timevalue < 30) {
            hickory.setAttribute("display", "");
            hickory.setAttribute("transform", "translate(" + (600+scalefactor*3*-1 ) + ", -144 )");
        }

        if (timevalue > 30 && timevalue < 66) {
            dickory.setAttribute("display", "");
            dickory.setAttribute("transform", "translate(" + (-795+scalefactor*2) + ", 0 )");
        }
        if (timevalue > 66) {
            dock.setAttribute("display", "");
            dock.setAttribute("transform", "translate(" + (1450+scalefactor*2*-1) + ", 144 )");
        }

        // Call ShowAndGrowElement again <timer_increment> milliseconds later.
        setTimeout("ShowAndGrowElement()", timer_increment)
    }
    window.ShowAndGrowElement = ShowAndGrowElement
  ]]></script>

  <rect
    fill="#2e3436"
    fill-rule="nonzero"
    stroke-width="3"
    y="0"
    x="0"
    height="600"
    width="800"
    id="rect3590"/>

    <text
       style="font-size:144px;font-style:normal;font-variant:normal;font-weight:bold;font-stretch:normal;fill:#000000;fill-opacity:1;stroke:none;font-family:Bitstream Vera Sans;-inkscape-font-specification:Bitstream Vera Sans Bold"
       x="50"
       y="350"
       id="hickory"
       display="none">
        Hickory,</text>
    <text
       style="font-size:144px;font-style:normal;font-variant:normal;font-weight:bold;font-stretch:normal;fill:#000000;fill-opacity:1;stroke:none;font-family:Bitstream Vera Sans;-inkscape-font-specification:Bitstream Vera Sans Bold"
       x="50"
       y="350"
       id="dickory"
       display="none">
        dickory,</text>
    <text
       style="font-size:144px;font-style:normal;font-variant:normal;font-weight:bold;font-stretch:normal;fill:#000000;fill-opacity:1;stroke:none;font-family:Bitstream Vera Sans;-inkscape-font-specification:Bitstream Vera Sans Bold"
       x="50"
       y="350"
       id="dock"
       display="none">
        dock!</text>
</svg>