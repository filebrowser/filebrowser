<html>
<body>
    <script>
        var x = "abc";
        function y {
        }
    </script>
    <style>
        .class {
            background: #124356;
        }
    </style>

    <p>
        Today's date: <%= (new java.util.Date()).toLocaleString()%>
    </p>
    <%! int i = 0; %>
    <jsp:declaration>
       int j = 10;
    </jsp:declaration>

    <%-- This is JSP comment --%>
    <%@ directive attribute="value" %>

    <h2>Select Languages:</h2>

    <form ACTION="jspCheckBox.jsp">
        <input type="checkbox" name="id" value="Java"> Java<BR>
        <input type="checkbox" name="id" value=".NET"> .NET<BR>
        <input type="checkbox" name="id" value="PHP"> PHP<BR>
        <input type="checkbox" name="id" value="C/C++"> C/C++<BR>
        <input type="checkbox" name="id" value="PERL"> PERL <BR>
        <input type="submit" value="Submit">
    </form>

    <%
    String select[] = request.getParameterValues("id"); 
    if (select != null && select.length != 0) {
        out.println("You have selected: ");
        for (int i = 0; i < select.length; i++) {
           out.println(select[i]); 
        }
    }
    %>
</body>
</html>