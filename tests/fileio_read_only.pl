my $fh;
open($fh, "<", "output.txt");
my $line1 = <$fh>;
say "Got: $line1";
close($fh);