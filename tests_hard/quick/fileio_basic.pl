# tests/quick/fileio_basic.pl
# Basic file I/O test

# Test 1: Write to file
say "Test 1: Basic write";
open(my $fh, ">", "test_output.txt");
print $fh "Hello from Perl!\n";
print $fh "Line 2\n";
say $fh "Line 3";
close($fh);
say "PASS: File written";

# Test 2: Read from file
say "Test 2: Basic read";
open($fh, "<", "test_output.txt");
my $line1 = <$fh>;
my $line2 = <$fh>;
my $line3 = <$fh>;
close($fh);

chomp($line1);
chomp($line2);
chomp($line3);

if ($line1 eq "Hello from Perl!") {
    say "PASS: Line 1 correct";
} else {
    say "FAIL: Line 1 wrong: $line1";
}

if ($line2 eq "Line 2") {
    say "PASS: Line 2 correct";
} else {
    say "FAIL: Line 2 wrong: $line2";
}

if ($line3 eq "Line 3") {
    say "PASS: Line 3 correct";
} else {
    say "FAIL: Line 3 wrong: $line3";
}

# Cleanup
unlink("test_output.txt");
say "Done!";
