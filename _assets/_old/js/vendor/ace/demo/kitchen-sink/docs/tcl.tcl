
proc dijkstra {graph origin} {
    # Initialize
    dict for {vertex distmap} $graph {
	dict set dist $vertex Inf
	dict set path $vertex {}
    }
    dict set dist $origin 0
    dict set path $origin [list $origin]
 
    while {[dict size $graph]} {
	# Find unhandled node with least weight
	set d Inf
	dict for {uu -} $graph {
	    if {$d > [set dd [dict get $dist $uu]]} {
		set u $uu
		set d $dd
	    }
	}
 
	# No such node; graph must be disconnected
	if {$d == Inf} break
 
	# Update the weights for nodes\
	 lead to by the node we've picked
	dict for {v dd} [dict get $graph $u] {
	    if {[dict exists $graph $v]} {
		set alt [expr {$d + $dd}]
		if {$alt < [dict get $dist $v]} {
		    dict set dist $v $alt
		    dict set path $v [list {*}[dict get $path $u] $v]
		}
	    }
	}
 
	# Remove chosen node from graph still to be handled
	dict unset graph $u
    }
    return [list $dist $path]
}