/*
    Testing syntax highlighting
    of Ace Editor
    for the Linden Scripting Language
*/

integer someIntNormal       = 3672;
integer someIntHex          = 0x00000000;
integer someIntMath         = PI_BY_TWO;

integer event               = 5673;                                             // invalid.illegal

key someKeyTexture          = TEXTURE_DEFAULT;
string someStringSpecial    = EOF;

some_user_defined_function_without_return_type(string inputAsString)
{
    llSay(PUBLIC_CHANNEL, inputAsString);
}

string user_defined_function_returning_a_string(key inputAsKey)
{
    return (string)inputAsKey;
}

default
{
    state_entry()
    {
        key someKey = NULL_KEY;
        someKey = llGetOwner();

        string someString = user_defined_function_returning_a_string(someKey);

        some_user_defined_function_without_return_type(someString);
    }

    touch_start(integer num_detected)
    {
        list agentsInRegion = llGetAgentList(AGENT_LIST_REGION, []);
        integer numOfAgents = llGetListLength(agentsInRegion);

        integer index;                                                          // defaults to 0
        for (; index <= numOfAgents - 1; index++)                               // for each agent in region
        {
            llRegionSayTo(llList2Key(agentsInRegion, index), PUBLIC_CHANNEL, "Hello, Avatar!");
        }
    }

    touch_end(integer num_detected)
    {
        someIntNormal       = 3672;
        someIntHex          = 0x00000000;
        someIntMath         = PI_BY_TWO;

        event               = 5673;                                             // invalid.illegal

        someKeyTexture      = TEXTURE_DEFAULT;
        someStringSpecial   = EOF;

        llSetInventoryPermMask("some item", MASK_NEXT, PERM_ALL);               // reserved.godmode

        llWhisper(PUBLIC_CHANNEL, "Leaving \"default\" now...");
        state other;
    }
}

state other
{
    state_entry()
    {
        llWhisper(PUBLIC_CHANNEL, "Entered \"state other\", returning to \"default\" again...");
        state default;
    }
}
