form Highlighter test
  sentence My_sentence This should all be a string
  text My_text This should also all be a string
  word My_word Only the first word is a string, the rest is invalid
  boolean Binary 1
  boolean Text no
  boolean Quoted "yes"
  comment This should be a string
  real left_Range -123.6
  positive right_Range_max 3.3
  integer Int 4
  natural Nat 4
endform

# External scripts
include /path/to/file
runScript: "/path/to/file"
execute /path/to/file

stopwatch

# old-style procedure call
call oldStyle "quoted" 2 unquoted string
assert oldStyle.local = 1

# New-style procedure call with parens
@newStyle("quoted", 2, "quoted string")
if praatVersion >= 5364 
  # New-style procedure call with colon
  @newStyle: "quoted", 2, "quoted string"
endif

# if-block with built-in variables
if windows
  # We are on Windows
elsif unix = 1 or !macintosh
  exitScript: "We are on Linux"
else macintosh == 1
  exit We are on Mac
endif

# inline if with inline comment
var = if macintosh = 1 then 0 else 1 fi ; This is an inline comment

# for-loop with explicit from using local variable
# and paren-style function calls and variable interpolation
n = numberOfSelected("Sound")
for i from newStyle.local to n
  sound'i' = selected("Sound", i)
  sound[i] = sound'i'
endfor

for i from 1 to n
  # Different styles of object selection
  select sound'i'
  sound = selected()
  sound$ = selected$("Sound")
  select Sound 'sound$'
  selectObject(sound[i])
  selectObject: sound
  
  # Pause commands
  beginPause("Viewing " + sound$)
  if i > 1
    button = endPause("Stop", "Previous",
      ...if i = total_sounds then "Finish" else "Next" fi,
      ...3, 1)
  else
    button = endPause("Stop",
      ...if i = total_sounds then "Finish" else "Next" fi,
      ...2, 1)  
  endif
  editor_name$ = if total_textgrids then "TextGrid " else "Sound " fi + name$
  nocheck editor 'editor_name$'
    nocheck Close
  nocheck endeditor
  
  # New-style standalone command call
  Rename: "SomeName"

  # Command call with assignment
  duration = Get total duration
  
  # Multi-line command with modifier
  pitch = noprogress To Pitch (ac): 0, 75, 15, "no",
    ...0.03, 0.45, 0.01, 0.35, 0.14, 600
    
  # do-style command with assignment
  minimum = do("Get minimum...", 0, 0, "Hertz", "Parabolic")

  # New-style multi-line command call with broken strings
  table = Create Table with column names: "table", 0,
    ..."file subject speaker
    ...f0 f1 f2 f3 " +
    ..."duration response"
  
  removeObject: pitch, table
    
  # Picture window commands
  selectObject: sound
  # do-style command
  do("Select inner viewport...", 1, 6, 0.5, 1.5)
  Black
  Draw... 0 0 0 0 "no" Curve
  Draw inner box
  Text bottom: "yes", sound$
  Erase all
  
  # Demo window commands
  demo Erase all
  demo Select inner viewport... 0 100 0 100
  demo Axes... 0 100 0 100
  demo Paint rectangle... white 0 100 0 100
  demo Text... 50 centre 50 half Click to finish
  demoWaitForInput ( )
  demo Erase all
  demo Text: 50, "centre", 50, "half", "Finished"
endfor

# An old-style sendpraat block
sendpraat Praat
  ...'newline$' Create Sound as pure tone... "tone" 1 0 0.4 44100 440 0.2 0.01 0.01
  ...'newline$' Play
  ...'newline$' Remove

# A new-style sendpraat block
beginSendPraat: "Praat"
  Create Sound as pure tone: "tone", 1, 0, 0.4, 44100, 440, 0.2, 0.01, 0.01
  duration = Get total duration
  Remove
endSendPraat: "duration"
appendInfoLine: "The generated sound lasted for ", duration, "seconds"

time = stopwatch
clearinfo
echo This script took 
print 'time' seconds to 
printline execute.

# Old-style procedure declaration
procedure oldStyle .str1$ .num .str2$
  .local = 1
endproc

# New-style procedure declaration
procedure newStyle (.str1$, .num, .str2$)
  .local = 1
endproc
