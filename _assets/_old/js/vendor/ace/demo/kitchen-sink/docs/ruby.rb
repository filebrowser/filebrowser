#!/usr/bin/ruby

# Program to find the factorial of a number
def fact(n)
    if n == 0
        1
    else
        n * fact(n-1)
    end
end

puts fact(ARGV[0].to_i)

class Range
  def to_json(*a)
    {
      'json_class'   => self.class.name, # = 'Range'
      'data'         => [ first, last, exclude_end? ]
    }.to_json(*a)
  end
end

{:id => ?", :key => "value"}


    herDocs = [<<'FOO', <<BAR, <<-BAZ, <<-`EXEC`] #comment
  FOO #{literal}
FOO
  BAR #{fact(10)}
BAR
  BAZ indented
    BAZ
        echo hi
    EXEC
puts herDocs