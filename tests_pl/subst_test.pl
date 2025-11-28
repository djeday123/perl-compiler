my $str = "Hello World";
$str =~ s/World/Perl/;
say $str;

my $text = "foo bar foo baz foo";
$text =~ s/foo/XXX/g;
say $text;

my $name = "JOHN DOE";
$name =~ s/john/Jane/i;
say $name;