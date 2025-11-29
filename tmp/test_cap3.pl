my $str = "Name: Alice";
$str =~ /Name: (\w+)/;
say "after match: $1";
say $1;
