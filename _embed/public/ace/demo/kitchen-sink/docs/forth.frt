: HELLO  ( -- )  CR ." Hello, world!" ; 

HELLO <cr>
Hello, world!

: [CHAR]   CHAR  POSTPONE LITERAL ; IMMEDIATE

0 value ii        0 value jj
0 value KeyAddr   0 value KeyLen
create SArray   256 allot   \ state array of 256 bytes
: KeyArray      KeyLen mod   KeyAddr ;

: get_byte      + c@ ;
: set_byte      + c! ;
: as_byte       255 and ;
: reset_ij      0 TO ii   0 TO jj ;
: i_update      1 +   as_byte TO ii ;
: j_update      ii SArray get_byte +   as_byte TO jj ;
: swap_s_ij
    jj SArray get_byte
       ii SArray get_byte  jj SArray set_byte
    ii SArray set_byte
;

: rc4_init ( KeyAddr KeyLen -- )
    256 min TO KeyLen   TO KeyAddr
    256 0 DO   i i SArray set_byte   LOOP
    reset_ij
    BEGIN
        ii KeyArray get_byte   jj +  j_update
        swap_s_ij
        ii 255 < WHILE
        ii i_update
    REPEAT
    reset_ij
;
: rc4_byte
    ii i_update   jj j_update
    swap_s_ij
    ii SArray get_byte   jj SArray get_byte +   as_byte SArray get_byte  xor
;