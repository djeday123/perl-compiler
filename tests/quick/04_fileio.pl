# Test file I/O

say "Testing file I/O...";

my $fh;
open($fh, ">", "test_file.txt");
print $fh "Hello\n";
print $fh "World\n";
close($fh);

say "PASS: file written";

open($fh, "<", "test_file.txt");
my $line1 = <$fh>;
my $line2 = <$fh>;
close($fh);

# Remove newlines with regex
$line1 =~ s/\n$//;
$line2 =~ s/\n$//;

if ($line1 eq "Hello") {
    say "PASS: read line 1";
} else {
    say "FAIL: line1 = $line1";
}

if ($line2 eq "World") {
    say "PASS: read line 2";
} else {
    say "FAIL: line2 = $line2";
}

say "Done!";
