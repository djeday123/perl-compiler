my $val = 42;
my $ref = \$val;
say "ref created";

my @arr = (10, 20, 30);
my $ref = \@arr;
print $ref->[1];
print "\n";
