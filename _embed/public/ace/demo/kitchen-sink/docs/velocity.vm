#*
  This is a sample comment block that
  spans multiple lines.
*#

#macro ( outputItem $item )
  <li>${item}</li>
#end

## Define the items to iterate
#set ( $items = [1, 2, 3, 4] )

<ul>
  ## Iterate over the items and output the evens.
  #foreach ( $item in $items )
    #if ( $_MathTool.mod($item, 2) == 0 )
      #outputItem ($item)
    #end
  #end
</ul>

<script>
  /*
    A sample function to decomstrate
    JavaScript highlighting and folding.
  */
  function foo(items, nada) {
    for (var i=0; i<items.length; i++) {
      alert(items[i] + "juhu\n");
    }
  }
</script>

<style>
  /*
    A sample style to decomstrate
    CSS highlighting and folding.
  */
  .class {
    font-family: Monaco, "Courier New", monospace;
    font-size: 12px;
    cursor: text;
  }
</style>