open(my $fh, "<", "test_input.txt");
my $line1 = <$fh>;
my $line2 = <$fh>;
my $line3 = <$fh>;
close($fh);

say "Line 1: $line1";
say "Line 2: $line2";
say "Line 3: $line3";