%{
   %{
      Ace Matlab demo
   %}
%}

classdef hello
   methods
      function greet(this)
         disp('Hello!')  % say hi
      end
   end
end

% transpose 
a = [ 'x''y', "x\n\
      y", 1' ]' + 2'