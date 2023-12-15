//!OpenSCAD

quality=128;
tubeHeight = 10;     //20
baseThickness =5;   // 5

//=======================
module boltHole(){
    cylinder(h=20,d=3.6,$fn=quality);
}

//=======================
module fixings(){
    translate([7,7,-1]){
        boltHole();
    }

    translate([7,48,-1]){
        boltHole();
    }
    
    //USB
    translate([-1,25,10]){
        cube([8,11,8]);
    }
    
    //12v power
    translate([-1,15,20]){
        rotate([0,90,0]){
            cylinder(h=10,d=6,$fn=quality);
        }
    }

}


//=======================
module Lid(){

    difference(){
        cube([81,56,3]);
        // hole for heatsink
        translate([61,25,-1]){
            cube([14,15,6]);
        }
        // hole for motor cable
        translate([61,15,-1]){
            cube([12,5,6]);
        }
        
    }
    //corner blocks
    translate([3,3,-4]){
        cube([5,5,5]);
    }
    translate([3,48,-4]){
        cube([5,5,5]);
    }
    translate([73,3,-4]){
        cube([5,5,5]);
    }
    translate([73,48,-4]){
        cube([5,5,5]);
    }


}

//=======================
module Base(){
    difference(){
        cube([81,56,30]);
        translate([3,3,3]){
            cube([75,50,40]);
        }

    }
    //corner ledges
    //bottom left
    translate([3,3,3]){
        cube([8,8,5]);
    }
    //top left
    translate([3,44,3]){
        cube([8,8,5]);
    }
    //bottom right
    translate([73,18,3]){
        cube([5,5,5]);
    }
    //top right
    translate([73,41,3]){
        cube([5,5,5]);
    }
     
}

printRotation = 180;
// ---------- Lid ----------
translate([0,0,33]){
    rotate([printRotation,0,0]){
        Lid();
    }
}


// ---------- Base part ----------
/*
difference(){
    Base();
    fixings();
}
*/
