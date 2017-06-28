<#ftl encoding="utf-8" />
<#setting locale="en_US" />
<#import "library" as lib />
<#--
    FreeMarker comment
    ${abc} <#assign a=12 />
-->

<!DOCTYPE html>
<html lang="en-us">
    <head>
        <meta charset="utf-8" />
        
        <title>${title!"FreeMarker"}<title>
    </head>
    
    <body>
    
        <h1>Hello ${name!""}</h1>
        
        <p>Today is: ${.now?date}</p>
        
        <#assign x = 13>
        <#if x &gt; 12 && x lt 14>x equals 13: ${x}</#if>
        
        <ul>
            <#list items as item>
                <li>${item_index}: ${item.name!?split("\n")[0]}</li>
            </#list>
        </ul>
        
        User directive: <@lib.function attr1=true attr2='value' attr3=-42.12>Test</@lib.function>
        <@anotherOne />
        
        <#if variable?exists>
            Deprecated
        <#elseif variable??>
            Better
        <#else>
            Default
        </#if>
        
        <img src="images/${user.id}.png" />
        
    </body>
</html>
