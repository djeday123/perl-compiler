my $val = 42;
my $ref = \$val;
say $$ref;
$$ref = 100;
say $val;
