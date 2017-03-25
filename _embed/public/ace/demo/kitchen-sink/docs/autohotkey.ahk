#NoEnv
SetBatchLines -1

CoordMode Mouse, Screen
OnExit GuiClose

zoom := 9

computeSize(){
	global as_x
	as_x := Round(ws_x/zoom/2 - 0.5)
	if (zoom>1) {
		pix := Round(zoom)
	} ele {
		pix := 1
	}
    ToolTip Message %as_x% %zoom% %ws_x% %hws_x% 
}

hdc_frame := DllCall("GetDC", UInt, MagnifierID)

; comment
DrawCross(byRef x="", rX,rY,z, dc){
        ;specify the style, thickness and color of the cross lines
    h_pen := DllCall( "gdi32.dll\CreatePen", Int, 0, Int, 1, UInt, 0x0000FF)
}

;Ctrl ^; Shift +; Win #; Alt !
^NumPadAdd::
^WheelUp::   
^;::   ;comment
    If(zoom < ws_x and ( A_ThisHotKey = "^WheelUp" or A_ThisHotKey ="^NumPadAdd") )
		zoom *= 1.189207115         ; sqrt(sqrt(2))
	Gosub,setZoom
return
