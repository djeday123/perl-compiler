my $val = 42;
my $ref = \$val;
my $x = $$ref;
say $x;
