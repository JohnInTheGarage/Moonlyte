//!OpenSCAD

/* 
The individual parts can be printed separately by commenting the call to the "assembly()" module, and 
instead adding calls to the relevant parts, i.e. steppersuppport() or mainbodytop() etc.
Some are better printed upside-down.
*/

q=120;      // quality number for cylinders
test=false; // alows looking inside
xpl=0;      // "explodes" the assembly vertically if > zero

//----------------------------
module Generic_Ring(outer, inner, height){
    difference(){
        cylinder(h=height, d1=outer, d2=outer, center=false, $fn = q );
        translate([0,0,-1]){
            cylinder(h=height+2, d1=inner, d2=inner, center=false, $fn = q );
        }
    }
}

//----------------------------
module cornerpillar(){
    difference(){
        cylinder(h=20,d=8, $fn=q);
        translate([0,0,-1]){
            cylinder(h=25,d=4, $fn=q);
        }
    }
}

//----------------------------
//square body
module bodysection(){
    difference(){
        cube([60,60,30], center=true);
        // main cavity
        translate([0,0,5]){
            cube([52,52,33], center=true);
        }
        //shaft hole
        translate([0,0,-19]){
            cylinder(h=10,d=24,$fn=q);
        }

        //corner access to bolts
        translate([-32,-32,-25]){
            cube([13,13,20]);
        }
        translate([-32,19,-25]){
            cube([13,13,20]);
        }
        translate([19,-32,-25]){
            cube([13,13,20]);
        }
        translate([19,19,-25]){
            cube([13,13,20]);
        }

        if (test){  // cut in half to show inside
            translate([-40,-31,-20]){
                cube([80,30,100]);
            }
        }

    }

    for (pillar = [45:90:360]) {
        rotate([0,0,pillar]){
            translate([33,0,0-5]){
                cornerpillar();
            }
        }    
    }
    
}

//----------------------------
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
module mainbodytop(){ 
    difference(){
        mirror([0,0,1]){
            bodysection();
        }
        stepperboltholes();
    }
    translate([26,-5,-17]){
        flange();
    }
    
    translate([-26,5,-17]){
        rotate([0,0,180]){
            flange();
        }
    }
    translate([-5,-26,-17]){
        rotate([0,0,270]){
            flange();
        }
    }
    translate([5,26,-17]){
        rotate([0,0,90]){
            flange();
        }
    }

}


//----------------------------
module mainbodybase(){ 
    module threeholes(){
        for (screw = [0:120:360]) {
            rotate([0,0,screw]){
                translate([16,0,-8]){
                    cylinder(h=20,d=3, $fn=q);
                }
            }
        }

    }
    module threepads(){
        for (screw = [0:120:360]) {
            rotate([0,0,screw]){
                translate([16,0,-14]){
                    cylinder(h=10,d=8, $fn=q);
                }
            }
        }
    }

    //-------------------
    difference(){
        union(){
            bodysection();
            translate([0,0,-18]){
                Generic_Ring(42, 24, 3);    
            }
            threepads();
            
        }    
        translate([0,0,-15]){
            threeholes();
        } 

    }
    
}


// attachment to rubber shaft
//----------------------------
module focusattachment(){

grooveheight=15;
    
module blobs(){
    for (side = [1:1:8]) {
        rotate([0,0,45*side]){
            translate([8,0,0]){
                cylinder(h=grooveheight, d=6, $fn=q);
            }
        }
    }
    cylinder(h=grooveheight, d=20, $fn=q);
    translate([0,0,14]){
        scale([1,1,0.5]){
            sphere(11);
        }
    }
}

    difference(){
        translate([0,0,1]){
            cylinder(h=21,d=26, $fn=q);
        }
        blobs();
    }
    translate([0,0,22]){
        cylinder(h=10,d=8, $fn=q);
    }
}

// Stepper shaft adapter
//----------------------------
module steppershaftadapter(){
    color("silver"){
        cylinder(h=23, d=18, $fn=q);
    }
}

//----------------------------
module steppermotor(){
    color("silver"){
        cylinder(h=16, d=5, $fn=q);
        translate([0,0,16]){
            cylinder(h=2, d=22, $fn=q);
        }
    }
    color("black"){
        translate([0,0,18+11.5]){
            cube([41,41,23],center=true);
        }
    }
    
}

//----------------------------
module steppersuppport(){

    difference(){
        cube([41,41,16],center=true);
        cube([25,25,25],center=true);
        stepperboltholes();
        // arches
        translate([0,0,-7.5]){
            rotate([90,0,0]){
                cylinder(h=50,d=20, $fn=60,center=true);
                rotate([0,90,0]){
                    cylinder(h=50,d=20, $fn=60,center=true);
                }
            }
        }
        
    }
    
}
//========================================
//========================================

module assembly(){

    translate([0,0,60+xpl*3]){
        steppermotor();         // For reference only, not printed
    }
    
    translate([0,0,63+7.5+xpl*2]){
        steppersuppport();
    }
    
    translate([0,0,48+xpl]){ 
        mainbodytop();
    }
    
    translate([0,0,18]){ 
        mainbodybase();
    }


    translate([0,0,47+xpl*0.8]){
        steppershaftadapter();      // For reference only, not printed
    }

    translate([0,0,23+xpl*0.6]){
        focusattachment();
    }

    //rubber focusing shaft
    translate([0,0,xpl*0.3]){
        color("black"){
            cylinder(h=35,d=21, $fn = 8);  // For reference only, not printed
        }
    }
    
}

assembly();

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