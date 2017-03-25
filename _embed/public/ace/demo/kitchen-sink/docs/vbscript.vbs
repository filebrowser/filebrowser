myfilename = "C:\Wikipedia - VBScript - Example - Hello World.txt"
MakeHelloWorldFile myfilename
 
Sub MakeHelloWorldFile (FileName)
'Create a new file in C: drive or overwrite existing file
   Set FSO = CreateObject("Scripting.FileSystemObject")
   If FSO.FileExists(FileName) Then 
      Answer = MsgBox ("File " & FileName & " exists ... OK to overwrite?", vbOKCancel)
      'If button selected is not OK, then quit now
      'vbOK is a language constant
      If Answer <> vbOK Then Exit Sub
   Else
      'Confirm OK to create
      Answer = MsgBox ("File " & FileName & " ... OK to create?", vbOKCancel)
      If Answer <> vbOK Then Exit Sub
   End If
   'Create new file (or replace an existing file)
   Set FileObject = FSO.CreateTextFile (FileName)
   FileObject.WriteLine "Time ... " & Now()
   FileObject.WriteLine "Hello World"
   FileObject.Close()
   MsgBox "File " & FileName & " ... updated."
End Sub