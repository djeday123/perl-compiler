open(my $fh, ">", "output.txt");
print $fh "Hello from Perl!\n";
print $fh "Line 2\n";
say $fh "Line 3";
close($fh);

say "File written!";

# Read it back
open(my $fh2, "<", "output.txt");
my $line1 = <$fh2>;
my $line2 = <$fh2>;
my $line3 = <$fh2>;
close($fh2);

say "Read back:";
say "1: $line1";
say "2: $line2";
say "3: $line3";