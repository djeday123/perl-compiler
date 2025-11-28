# Write to file
open(my $fh, ">", "test_output.txt");
print "Writing to file...\n";
close($fh);

say "File written.";

# Read from file
open(my $fh2, "<", "test_output.txt");
my $line = <$fh2>;
say "Read: $line";
close($fh2);

say "Done!";