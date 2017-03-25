for op = (:+, :*, :&, :|, :$)
  @eval ($op)(a,b,c) = ($op)(($op)(a,b),c)
end

v = α';
function g(x,y)
  return x * y
  x + y
end

cd("data") do
    open("outfile", "w") do f
        write(f, data)
    end
end
