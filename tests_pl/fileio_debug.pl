say "Starting...";
my $result = open(my $fh, "<", "test_input.txt");
say "Open result: $result";
say "FH value: $fh";
my $line = <$fh>;
say "Line value: $line";
say "Done";