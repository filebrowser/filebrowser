class Haxe 
{
    public static function main() 
    {
        // Say Hello!
        var greeting:String = "Hello World";
        trace(greeting);
        
        var targets:Array<String> = ["Flash","Javascript","PHP","Neko","C++","iOS","Android","webOS"];
        trace("Haxe is a great language that can target:");
        for (target in targets)
        {
            trace (" - " + target);
        }
        trace("And many more!");
    }
}