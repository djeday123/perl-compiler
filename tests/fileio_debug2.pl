say "Opening file...";
my $result = open(my $fh, ">", "output.txt");
say "Open result: $result";
say "FH name: $fh";
print $fh "Test line\n";
say "After print";
close($fh);
say "Closed";