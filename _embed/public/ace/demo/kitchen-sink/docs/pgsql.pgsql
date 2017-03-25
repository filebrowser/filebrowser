
BEGIN;

/**
* Samples from PostgreSQL src/tutorial/basics.source
*/
CREATE TABLE weather (
	city		varchar(80),
	temp_lo		int,		-- low temperature
	temp_hi		int,		-- high temperature
	prcp		real,		-- precipitation
	"date"		date
);

CREATE TABLE cities (
	name		varchar(80),
	location	point
);


INSERT INTO weather
    VALUES ('San Francisco', 46, 50, 0.25, '1994-11-27');

INSERT INTO cities
    VALUES ('San Francisco', '(-194.0, 53.0)');

INSERT INTO weather (city, temp_lo, temp_hi, prcp, "date")
    VALUES ('San Francisco', 43, 57, 0.0, '1994-11-29');

INSERT INTO weather (date, city, temp_hi, temp_lo)
    VALUES ('1994-11-29', 'Hayward', 54, 37);


SELECT city, (temp_hi+temp_lo)/2 AS temp_avg, "date" FROM weather;

SELECT city, temp_lo, temp_hi, prcp, "date", location
    FROM weather, cities
    WHERE city = name;



/**
* Dollar quotes starting at the end of the line are colored as SQL unless
* a special language tag is used. Dollar quote syntax coloring is implemented
* for Perl, Python, JavaScript, and Json.
*/
create or replace function blob_content_chunked(
    in p_data bytea, 
    in p_chunk integer)
returns setof bytea as $$
-- Still SQL comments
declare
	v_size integer = octet_length(p_data);
begin
	for i in 1..v_size by p_chunk loop
		return next substring(p_data from i for p_chunk);
	end loop;
end;
$$ language plpgsql stable;


-- pl/perl
CREATE FUNCTION perl_max (integer, integer) RETURNS integer AS $perl$
    # perl comment...
    my ($x,$y) = @_;
    if (! defined $x) {
        if (! defined $y) { return undef; }
        return $y;
    }
    if (! defined $y) { return $x; }
    if ($x > $y) { return $x; }
    return $y;
$perl$ LANGUAGE plperl;

-- pl/python
CREATE FUNCTION usesavedplan() RETURNS trigger AS $python$
    # python comment...
    if SD.has_key("plan"):
        plan = SD["plan"]
    else:
        plan = plpy.prepare("SELECT 1")
        SD["plan"] = plan
$python$ LANGUAGE plpythonu;

-- pl/v8 (javascript)
CREATE FUNCTION plv8_test(keys text[], vals text[]) RETURNS text AS $javascript$
var o = {};
for(var i=0; i<keys.length; i++){
 o[keys[i]] = vals[i];
}
return JSON.stringify(o);
$javascript$ LANGUAGE plv8 IMMUTABLE STRICT;

-- json
select * from json_object_keys($json$
{
  "f1": 5,
  "f2": "test",
  "f3": {}
}
$json$);


-- psql commands
\df cash*


-- Some string samples.
select 'don''t do it now;' || 'maybe later';
select E'dont\'t do it';
select length('some other''s stuff' || $$cat in hat's stuff $$);

select $$ strings
over multiple 
lines - use dollar quotes
$$;

END;
