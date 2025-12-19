my @arr = (10, 20, 30);
my $arr_ref = \@arr;
say $arr_ref->[1];

my %h = ("x" => 5, "y" => 10);
my $h_ref = \%h;
say $h_ref->{x};
