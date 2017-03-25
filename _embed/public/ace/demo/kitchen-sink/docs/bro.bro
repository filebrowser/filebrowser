##! Add countries for the originator and responder of a connection
##! to the connection logs.

module Conn;

export {
	redef record Conn::Info += {
		## Country code for the originator of the connection based 
		## on a GeoIP lookup.
		orig_cc: string &optional &log;
		## Country code for the responser of the connection based 
		## on a GeoIP lookup.
		resp_cc: string &optional &log;
	};
}

event connection_state_remove(c: connection) 
	{
	local orig_loc = lookup_location(c$id$orig_h);
	if ( orig_loc?$country_code )
		c$conn$orig_cc = orig_loc$country_code;

	local resp_loc = lookup_location(c$id$resp_h);
	if ( resp_loc?$country_code )
		c$conn$resp_cc = resp_loc$country_code;
	}