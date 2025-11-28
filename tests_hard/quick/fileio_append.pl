# tests/quick/fileio_append.pl
# Test append mode

say "Testing append mode...";

# Write initial content
open(my $fh, ">", "append_test.txt");
print $fh "Line 1\n";
close($fh);

# Append more content
open($fh, ">>", "append_test.txt");
print $fh "Line 2\n";
print $fh "Line 3\n";
close($fh);

# Read and verify
open($fh, "<", "append_test.txt");
my @lines;
my $line = <$fh>;
while (defined $line) {
    chomp($line);
    push(@lines, $line);
    $line = <$fh>;
}
close($fh);

if (scalar(@lines) == 3) {
    say "PASS: Got 3 lines";
} else {
    say "FAIL: Expected 3 lines, got " . scalar(@lines);
}

if ($lines[0] eq "Line 1" && $lines[1] eq "Line 2" && $lines[2] eq "Line 3") {
    say "PASS: Content is correct";
} else {
    say "FAIL: Content mismatch";
    foreach my $i (0..$#lines) {
        say "  [$i]: $lines[$i]";
    }
}

# Cleanup
unlink("append_test.txt");
say "Done!";
