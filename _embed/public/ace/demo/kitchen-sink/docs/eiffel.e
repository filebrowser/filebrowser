note
	description: "Represents a person."

class
	PERSON

create
	make, make_unknown

feature {NONE} -- Creation

	make (a_name: like name)
			-- Create a person with `a_name' as `name'.
		do
			name := a_name
		ensure
			name = a_name
		end

	make_unknown
		do ensure
			name = Void
		end

feature -- Access

	name: detachable STRING
			-- Full name or Void if unknown.

end