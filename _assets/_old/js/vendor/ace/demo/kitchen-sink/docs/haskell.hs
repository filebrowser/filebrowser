-- Type annotation (optional)
fib :: Int -> Integer
 
-- With self-referencing data
fib n = fibs !! n
        where fibs = 0 : scanl (+) 1 fibs
        -- 0,1,1,2,3,5,...
 
-- Same, coded directly
fib n = fibs !! n
        where fibs = 0 : 1 : next fibs
              next (a : t@(b:_)) = (a+b) : next t
 
-- Similar idea, using zipWith
fib n = fibs !! n
        where fibs = 0 : 1 : zipWith (+) fibs (tail fibs)
 
-- Using a generator function
fib n = fibs (0,1) !! n
        where fibs (a,b) = a : fibs (b,a+b)