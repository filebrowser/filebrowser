define("ace/snippets/perl",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "# #!/usr/bin/perl\n\
snippet #!\n\
	#!/usr/bin/env perl\n\
\n\
# Hash Pointer\n\
snippet .\n\
	 =>\n\
# Function\n\
snippet sub\n\
	sub ${1:function_name} {\n\
		${2:#body ...}\n\
	}\n\
# Conditional\n\
snippet if\n\
	if (${1}) {\n\
		${2:# body...}\n\
	}\n\
# Conditional if..else\n\
snippet ife\n\
	if (${1}) {\n\
		${2:# body...}\n\
	}\n\
	else {\n\
		${3:# else...}\n\
	}\n\
# Conditional if..elsif..else\n\
snippet ifee\n\
	if (${1}) {\n\
		${2:# body...}\n\
	}\n\
	elsif (${3}) {\n\
		${4:# elsif...}\n\
	}\n\
	else {\n\
		${5:# else...}\n\
	}\n\
# Conditional One-line\n\
snippet xif\n\
	${1:expression} if ${2:condition};${3}\n\
# Unless conditional\n\
snippet unless\n\
	unless (${1}) {\n\
		${2:# body...}\n\
	}\n\
# Unless conditional One-line\n\
snippet xunless\n\
	${1:expression} unless ${2:condition};${3}\n\
# Try/Except\n\
snippet eval\n\
	local $@;\n\
	eval {\n\
		${1:# do something risky...}\n\
	};\n\
	if (my $e = $@) {\n\
		${2:# handle failure...}\n\
	}\n\
# While Loop\n\
snippet wh\n\
	while (${1}) {\n\
		${2:# body...}\n\
	}\n\
# While Loop One-line\n\
snippet xwh\n\
	${1:expression} while ${2:condition};${3}\n\
# C-style For Loop\n\
snippet cfor\n\
	for (my $${2:var} = 0; $$2 < ${1:count}; $$2${3:++}) {\n\
		${4:# body...}\n\
	}\n\
# For loop one-line\n\
snippet xfor\n\
	${1:expression} for @${2:array};${3}\n\
# Foreach Loop\n\
snippet for\n\
	foreach my $${1:x} (@${2:array}) {\n\
		${3:# body...}\n\
	}\n\
# Foreach Loop One-line\n\
snippet fore\n\
	${1:expression} foreach @${2:array};${3}\n\
# Package\n\
snippet package\n\
	package ${1:`substitute(Filename('', 'Page Title'), '^.', '\\u&', '')`};\n\
\n\
	${2}\n\
\n\
	1;\n\
\n\
	__END__\n\
# Package syntax perl >= 5.14\n\
snippet packagev514\n\
	package ${1:`substitute(Filename('', 'Page Title'), '^.', '\\u&', '')`} ${2:0.99};\n\
\n\
	${3}\n\
\n\
	1;\n\
\n\
	__END__\n\
#moose\n\
snippet moose\n\
	use Moose;\n\
	use namespace::autoclean;\n\
	${1:#}BEGIN {extends '${2:ParentClass}'};\n\
\n\
	${3}\n\
# parent\n\
snippet parent\n\
	use parent qw(${1:Parent Class});\n\
# Read File\n\
snippet slurp\n\
	my $${1:var} = do { local $/; open my $file, '<', \"${2:file}\"; <$file> };\n\
	${3}\n\
# strict warnings\n\
snippet strwar\n\
	use strict;\n\
	use warnings;\n\
# older versioning with perlcritic bypass\n\
snippet vers\n\
	## no critic\n\
	our $VERSION = '${1:version}';\n\
	eval $VERSION;\n\
	## use critic\n\
# new 'switch' like feature\n\
snippet switch\n\
	use feature 'switch';\n\
\n\
# Anonymous subroutine\n\
snippet asub\n\
	sub {\n\
	 	${1:# body }\n\
	}\n\
\n\
\n\
\n\
# Begin block\n\
snippet begin\n\
	BEGIN {\n\
		${1:# begin body}\n\
	}\n\
\n\
# call package function with some parameter\n\
snippet pkgmv\n\
	__PACKAGE__->${1:package_method}(${2:var})\n\
\n\
# call package function without a parameter\n\
snippet pkgm\n\
	__PACKAGE__->${1:package_method}()\n\
\n\
# call package \"get_\" function without a parameter\n\
snippet pkget\n\
	__PACKAGE__->get_${1:package_method}()\n\
\n\
# call package function with a parameter\n\
snippet pkgetv\n\
	__PACKAGE__->get_${1:package_method}(${2:var})\n\
\n\
# complex regex\n\
snippet qrx\n\
	qr/\n\
	     ${1:regex}\n\
	/xms\n\
\n\
#simpler regex\n\
snippet qr/\n\
	qr/${1:regex}/x\n\
\n\
#given\n\
snippet given\n\
	given ($${1:var}) {\n\
		${2:# cases}\n\
		${3:# default}\n\
	}\n\
\n\
# switch-like case\n\
snippet when\n\
	when (${1:case}) {\n\
		${2:# body}\n\
	}\n\
\n\
# hash slice\n\
snippet hslice\n\
	@{ ${1:hash}  }{ ${2:array} }\n\
\n\
\n\
# map\n\
snippet map\n\
	map {  ${2: body }    }  ${1: @array } ;\n\
\n\
\n\
\n\
# Pod stub\n\
snippet ppod\n\
	=head1 NAME\n\
\n\
	${1:ClassName} - ${2:ShortDesc}\n\
\n\
	=head1 SYNOPSIS\n\
\n\
	  use $1;\n\
\n\
	  ${3:# synopsis...}\n\
\n\
	=head1 DESCRIPTION\n\
\n\
	${4:# longer description...}\n\
\n\
\n\
	=head1 INTERFACE\n\
\n\
\n\
	=head1 DEPENDENCIES\n\
\n\
\n\
	=head1 SEE ALSO\n\
\n\
\n\
# Heading for a subroutine stub\n\
snippet psub\n\
	=head2 ${1:MethodName}\n\
\n\
	${2:Summary....}\n\
\n\
# Heading for inline subroutine pod\n\
snippet psubi\n\
	=head2 ${1:MethodName}\n\
\n\
	${2:Summary...}\n\
\n\
\n\
	=cut\n\
# inline documented subroutine\n\
snippet subpod\n\
	=head2 $1\n\
\n\
	Summary of $1\n\
\n\
	=cut\n\
\n\
	sub ${1:subroutine_name} {\n\
		${2:# body...}\n\
	}\n\
# Subroutine signature\n\
snippet parg\n\
	=over 2\n\
\n\
	=item\n\
	Arguments\n\
\n\
\n\
	=over 3\n\
\n\
	=item\n\
	C<${1:DataStructure}>\n\
\n\
	  ${2:Sample}\n\
\n\
\n\
	=back\n\
\n\
\n\
	=item\n\
	Return\n\
\n\
	=over 3\n\
\n\
\n\
	=item\n\
	C<${3:...return data}>\n\
\n\
\n\
	=back\n\
\n\
\n\
	=back\n\
\n\
\n\
\n\
# Moose has\n\
snippet has\n\
	has ${1:attribute} => (\n\
		is	    => '${2:ro|rw}',\n\
		isa 	=> '${3:Str|Int|HashRef|ArrayRef|etc}',\n\
		default => sub {\n\
			${4:defaultvalue}\n\
		},\n\
		${5:# other attributes}\n\
	);\n\
\n\
\n\
# override\n\
snippet override\n\
	override ${1:attribute} => sub {\n\
		${2:# my $self = shift;};\n\
		${3:# my ($self, $args) = @_;};\n\
	};\n\
\n\
\n\
# use test classes\n\
snippet tuse\n\
	use Test::More;\n\
	use Test::Deep; # (); # uncomment to stop prototype errors\n\
	use Test::Exception;\n\
\n\
# local test lib\n\
snippet tlib\n\
	use lib qw{ ./t/lib };\n\
\n\
#test methods\n\
snippet tmeths\n\
	$ENV{TEST_METHOD} = '${1:regex}';\n\
\n\
# runtestclass\n\
snippet trunner\n\
	use ${1:test_class};\n\
	$1->runtests();\n\
\n\
# Test::Class-style test\n\
snippet tsub\n\
	sub t${1:number}_${2:test_case} :Test(${3:num_of_tests}) {\n\
		my $self = shift;\n\
		${4:#  body}\n\
\n\
	}\n\
\n\
# Test::Routine-style test\n\
snippet trsub\n\
	test ${1:test_name} => { description => '${2:Description of test.}'} => sub {\n\
		my ($self) = @_;\n\
		${3:# test code}\n\
	};\n\
\n\
#prep test method\n\
snippet tprep\n\
	sub prep${1:number}_${2:test_case} :Test(startup) {\n\
		my $self = shift;\n\
		${4:#  body}\n\
	}\n\
\n\
# cause failures to print stack trace\n\
snippet debug_trace\n\
	use Carp; # 'verbose';\n\
	# cloak \"die\"\n\
	# warn \"warning\"\n\
	$SIG{'__DIE__'} = sub {\n\
		require Carp; Carp::confess\n\
	};\n\
\n\
";
exports.scope = "perl";

});
