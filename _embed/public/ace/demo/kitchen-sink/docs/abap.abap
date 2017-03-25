***************************************
** Program: EXAMPLE                  **
** Author: Joe Byte, 07-Jul-2007     **
***************************************
 
REPORT BOOKINGS.
 
* Read flight bookings from the database
SELECT * FROM FLIGHTINFO
  WHERE CLASS = 'Y'       "Y = economy
  OR    CLASS = 'C'.      "C = business
(...)

REPORT TEST.
WRITE 'Hello World'.

USERPROMPT = 'Please double-click on a line in the output list ' &
             'to see the complete details of the transaction.'.


DATA LAST_EOM    TYPE D.  "last end-of-month date
 
* Start from today's date
  LAST_EOM = SY-DATUM.
* Set characters 6 and 7 (0-relative) of the YYYYMMDD string to "01",
* giving the first day of the current month
  LAST_EOM+6(2) = '01'.
* Subtract one day
  LAST_EOM = LAST_EOM - 1.
 
  WRITE: 'Last day of previous month was', LAST_EOM.
  
DATA : BEGIN OF I_VBRK OCCURS 0,
         VBELN LIKE VBRK-VBELN,
         ZUONR LIKE VBRK-ZUONR,
       END OF I_VBRK.

SORT i_vbrk BY vbeln ASCENDING.
SORT i_vbrk BY vbeln DESCENDING.

RETURN.