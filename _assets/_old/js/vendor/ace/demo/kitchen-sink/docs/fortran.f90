#include "globalDefines.h"

!=========================================================
program main
!========================================================= 
    use params_module, only : nx, ny, nz

    implicit none

    integer, parameter :: g = 9.81
    real, allocatable, dimension(:,:,:) :: array
    integer :: a, b, c
    real*8 :: x, y, z
      
    b = 5
    c = 7

#ifdef ARRAY_COMP
    allocate(array(10,10,10), status=a)

    write(c,'(i5.5)') b
#endif

    if(x.lt.5.0) then
        array(:,:,:) = g
    else
        array(:,:,:) = x - y
    endif

    return
!========================================================    
end program main
!========================================================
