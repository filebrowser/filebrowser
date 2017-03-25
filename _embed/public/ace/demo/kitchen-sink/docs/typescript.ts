class Greeter {
	greeting: string;
	constructor (message: string) {
		this.greeting = message;
	}
	greet() {
		return "Hello, " + this.greeting;
	}
}   

var greeter = new Greeter("world");

var button = document.createElement('button')
button.innerText = <string>"Say Hello";
button.onclick = function() {
	alert(greeter.greet())
}

document.body.appendChild(button)

class Snake extends Animal {
   move() {
       alert("Slithering...");
       super(5);
   }
}

class Horse extends Animal {
   move() {
       alert("Galloping...");
       super.move(45);
   }
}

module Sayings {
    export class Greeter {
        greeting: string;
        constructor (message: string) {
            this.greeting = message;
        }
        greet() {
            return "Hello, " + this.greeting;
        }
    }
}
module Mankala {
   export class Features {
       public turnContinues = false;
       public seedStoredCount = 0;
       public capturedCount = 0;
       public spaceCaptured = NoSpace;

       public clear() {
           this.turnContinues = false;
           this.seedStoredCount = 0;
           this.capturedCount = 0;
           this.spaceCaptured = NoSpace;
       }

       public toString() {
           var stringBuilder = "";
           if (this.turnContinues) {
               stringBuilder += " turn continues,";
           }
           stringBuilder += " stores " + this.seedStoredCount;
           if (this.capturedCount > 0) {
               stringBuilder += " captures " + this.capturedCount + " from space " + this.spaceCaptured;
           }
           return stringBuilder;
       }
   }
}