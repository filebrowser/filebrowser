// computes factorial of a number
factorial := method(n,
    if(n == 0, return 1)
    res := 1
    Range 1 to(n) foreach(i, res = res * i)
)