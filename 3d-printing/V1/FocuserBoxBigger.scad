//!OpenSCAD
/*
* 2024.01.13 Added an IR intervalometer for my EOS-M.
* 
* This bigger version of the box takes the same size circuit board 
* as the original but allows use of my old Canon EOS-M by including
* space for the optional manual focus buttons and the Intervalometer switch.
*/
quality=128;

//=======================
module Hole(diam){
    cylinder(h=20,d=diam,$fn=quality);
}

//=======================
module fixings(){
    translate([7,7,-1]){
        Hole(3.6);
    }

    translate([7,48,-1]){
        Hole(3.6);
    }
    
    //USB
    translate([-1,25,10]){
        cube([8,11,8]);
    }
    
    //12v power
    translate([-1,15,20]){
        rotate([0,90,0]){
            Hole(6);
        }
    }
    
    // holes for manual focus buttons
    translate([40,85,15]){
        rotate([90,0,0]){
            Hole(6.6);
        }
    }
    translate([60,85,15]){
        rotate([90,0,0]){
            Hole(6.6);
        }
    }
    
    // hole for intervalometer switch
    translate([-5,65,15]){
        rotate([0,90,0]){
            Hole(6.6);
        }
    }

}


//=======================
module Lid(){

    difference(){
        cube([81,80,3]);
        // hole for heatsink
        translate([61,25,-1]){
            cube([14,15,6]);
        }
        // hole for & intervalometer motor cables 
        translate([57,15,-1]){
            cube([18,5,6]);
        }
        
    }
    //corner blocks
    translate([3,3,-4]){
        cube([5,5,5]);
    }
    translate([3,72,-4]){
        cube([5,5,5]);
    }
    translate([73,3,-4]){
        cube([5,5,5]);
    }
    translate([73,72,-4]){
        cube([5,5,5]);
    }


}

//=======================
module Base(){
    difference(){
        cube([81,80,30]);
        translate([3,3,3]){
            // original cube([75,50,40]);
            cube([75,74,40]);
        }
    }
    //divider blocks same width as original box
    //to steady circuit board
    translate([0,53,3]){
        translate([3,0,0]){
            cube([5,3,8]);
        }
        translate([73,0,0]){
            cube([5,3,8]);
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

difference(){
    Base();
    fixings();
}

