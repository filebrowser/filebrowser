:: batch file highlighting in Ace!
@echo off

CALL set var1=%cd%
echo unhide everything in %var1%!

:: FOR loop in bat is super strange!
FOR /f "tokens=*" %%G IN ('dir /A:D /b') DO (
echo %var1%%%G
attrib -r -a -h -s "%var1%%%G" /D /S
)

pause

REM that's all
