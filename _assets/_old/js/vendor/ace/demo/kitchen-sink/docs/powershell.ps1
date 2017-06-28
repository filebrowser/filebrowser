# This is a simple comment
function Hello($name) {
  Write-host "Hello $name"
}

function add($left, $right=4) {
    if ($right -ne 4) {
        return $left
    } elseif ($left -eq $null -and $right -eq 2) {
        return 3
    } else {
        return 2
    }
}

$number = 1 + 2;
$number += 3

Write-Host Hello -name "World"

$an_array = @(1, 2, 3)
$a_hash = @{"something" = "something else"}

& notepad .\readme.md
