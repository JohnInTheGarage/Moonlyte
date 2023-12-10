//!OpenSCAD

include <Pulley_composer.scad>

/* 
The individual parts can be printed separately by commenting the call to the "assembly()" module, and 
instead adding calls to the relevant parts, i.e. steppersuppport() or mainbodytop() etc.
Some are better printed upside-down.
*/

q=120;      // quality number for cylinders
test=false; // alows looking inside
xpl=0;      // "explodes" the assembly vertically if > zero


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

//----------------------------
module steppermotor(){
    color("silver"){
        cylinder(h=16, d=5, $fn=q);
        cylinder(h=2, d=22, $fn=q);
    }
    color("black"){
        translate([0,0,-11.5]){
            cube([41,41,23],center=true);
        }
    }
    
}

//========================================
//========================================

module assembly(){

    translate([0,0,24+xpl*2]){
        steppermotor();         // For reference only, not printed
    }
    
    translate([-20,-23,0+xpl]){
        plate();                // For reference only, not printed
    }
    
    translate([0,60,-5-xpl]){
        cylinder(d=13.3,h=25, $fn=q);   // For reference only, not printed
    }
    
    translate([0,60,7-xpl]){
        // Calling external Pulley_composer.scad, tweaked from the thingyverse download
        pulley ( "GT2 2mm" , GT2_2mm_pulley_dia , 0.764 , 1.494 );
    }
    
    //translate([-15,20,-25-xpl]){
    //    cube(30);
    //}
}

assembly();

//translate([-19,-20,175]){
//    color("grey"){
//        cube([38, 150,2]);
//    }
//}
/* ------------ measurement objects ---------------
color("red"){
    //distance from top of stepper adapter to baseplate
    translate([10,0,0]){
        cube([2,2,70]);
    }
    //bolt length for attaching to scope
    translate([-10,-15,-10]){
        cube ([2,2,25]);
    }
    //bolt length for attaching to stepper
    translate([15,25,59]){
        cube ([2,2,25]);
    }
    translate([1,-15,10]){
        cube ([1,1,50]);
    }
}
//corner bolt
translate([-30,-30,6]){
    color("silver"){
        cylinder(h=50, d=4, $fn=q);
    }
}
*/