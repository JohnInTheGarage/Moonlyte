//!OpenSCAD

include <Pulley_composer.scad>

/* 
The individual parts can be printed separately by commenting the call to the "assembly()" module, and 
instead adding calls to the relevant parts, i.e. steppersuppport() or mainbodytop() etc.
Some are better printed upside-down.
*/

q=100;      // quality number for cylinders
test=false; // allows looking inside
xpl=0;      // "explodes" the assembly vertically if > zero

coverBigRadius = 22.5;
coverSmallRadius= 15;
stepperSize = 41.5;
printing = true;
// =========================================
module boltBlock(holeSize){
    difference(){
        cube([10,10,3], center=true);
        translate([0,0,-3]){
            cylinder(h=5,d=holeSize, $fn=q);
        }
    }
}

// =========================================
module beltcover(bigC, smallC){
    translate([0,0,-11]){
        linear_extrude(height = 25, center = false, convexity = 10, twist = 0, slices = 20, $fn =q) {
            hull() {
                translate([60,0,0]){ 
                    circle(bigC);
                }
                circle(smallC);
            }
        }
    }
}

// =========================================
module steppermotor(){
    color("silver"){
        cylinder(h=23, d=5, $fn=q);
        cylinder(h=2, d=22, $fn=q);
    }

    color("black"){
        difference(){
            translate([0,0,-11.5]){
                cube([stepperSize,stepperSize,23],center=true);
            }
            if (test){
                translate([-20,-10,-30]){
                   // cube([stepperSize,stepperSize,60]);
                }
            }                
        }
    }
    
}
module cover_ellipseHoles(){
    for(angle = [15 : 30 : 180]){
        rotate([90,0,angle]){
            scale([0.8,2,1]){
                cylinder(h=60,d=10,$fn=q);
            }
        }
    }
}

// =========================================
module cover_CurveUnderPulleyEnd(radius){
    translate([60,0,-35]){
        difference(){
            cylinder(r=radius,h=25,$fn=q);
            translate([0,0,-1]){
                cylinder(r=(radius - 1.5),h=35,$fn=q);
            }
            translate([-35,0,-3]){
                cube(60,center=true);
            }
            translate([0,0,12]){
                cover_ellipseHoles();
            }
            
        }
        
    }
}

// =========================================
module cover_StepperFixing(){
    
    difference(){
        cube([stepperSize +4.2, stepperSize +4.2, 8], center=true);
        translate([0,0,-2]){
            cube([stepperSize +0.2, stepperSize +0.2, 8], center=true);
        }
        translate([-13,-13,-5]){
            cube([40, 26, 12], center=false);
        }
        
//        if (test){
//            translate([-10,-10,-20]){
//                cube([stepperSize,stepperSize,30]);
//            }
//        }                

    }
    translate([0,-27,2.5]){
        boltBlock(4.1);
    }
    translate([0,27,2.5]){
        boltBlock(4.1);
    }

}


// =========================================
module cover(){
    
    difference(){
        beltcover(coverBigRadius, coverSmallRadius);
        translate([0,0,-1.5]){
            beltcover(coverBigRadius-1.5, coverSmallRadius-1.5);
        }
        translate([60,0,2]){
            cover_ellipseHoles();
        }

        if (test){
            translate([-15,0,-15]){
                cube([120,30,50]);
            }
        }
    }

    cover_CurveUnderPulleyEnd(coverBigRadius);
    
    //Bolt holes & their support
    translate([0,0,-9.5]){
        translate([0,-27,0]){
            boltBlock(4.1);
        }
        translate([0,27,0]){
            boltBlock(4.1);
        }
        difference(){
            cube([10,50,3], center=true);
            cube(26, center=true);
        }
    }
    
}


// =========================================
module plate(){
    //----------------------
    module stepperboltholes(){
        translate([15.5,15.5,-10]){
            cylinder(h=40,d=3,$fn=q);
        }
        translate([15.5,-15.5,-10]){
            cylinder(h=40,d=3,$fn=q);
        }
        translate([-15.5,15.5,-10]){
            cylinder(h=40,d=3,$fn=q);
        }
        translate([-15.5,-15.5,-10]){
            cylinder(h=40,d=3,$fn=q);
        }

    }
    //----------------------
    module pulleyholes(){
        for (screw = [90:120:360]) {
            rotate([0,0,screw]){
                translate([16,0,-8]){
                    cylinder(h=20,d=3, $fn=q);
                }
            }
        }
        cylinder(d=17,h=5, $fn=q);

    }


    color("grey"){
        difference(){
            cube([40,107,2]);
        
            translate([20,22.5,-1]){
                stepperboltholes();
            }
            translate([20,83,-1]){
                pulleyholes();
            }
        }
    }
}

//-----------------------------
module flange(){
    rotate([270,0,0]){
        difference(){
            scale([1,2,1]){
                cylinder(h=10, d=5, $fn=30);
            }
            translate([0,-15,-1]){
                cube(30);
            }
            translate([-10,0,-1]){
                cube(30);
            }
            
        }
    }
}

//========================================
//========================================
module assembly(){
if (!printing){
    translate([0,0,24+xpl*2]){
        steppermotor();         // For reference only, not printed
    }
    
    translate([-20,-23,0+xpl]){
        plate();                // For reference only, not printed
    }
    
    translate([0,60,-5-xpl]){
        cylinder(d=13.3,h=25, $fn=q);   // For reference only, not printed
    }
}    
    translate([0,60,7-xpl]){
        // Calling external Pulley_composer.scad, tweaked from the thingyverse download
        //pulley ( "GT2 2mm" , GT2_2mm_pulley_dia , 0.764 , 1.494 );
    }
    
    rotate([0,0,90]){
        translate([0,0,37+xpl]){
        rotate([0,0,0]){
//            cover();
        }
    
    
        translate([0,0,-15]){
            cover_StepperFixing();
        }
    }
}
   
    
}

assembly();
