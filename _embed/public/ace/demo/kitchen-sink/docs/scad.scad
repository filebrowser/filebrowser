// ace can highlight scad!
module Element(xpos, ypos, zpos){
	translate([xpos,ypos,zpos]){
		union(){
			cube([10,10,4],true);
			cylinder(10,15,5);
			translate([0,0,10])sphere(5);
		}
	}
}

union(){
	for(i=[0:30]){
		# Element(0,0,0);
		Element(15*i,0,0);
	}
}

for (i = [3, 5, 7, 11]){
	rotate([i*10,0,0])scale([1,1,i])cube(10);
}