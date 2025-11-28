# Test arrays

say "Testing arrays...";

my @arr = (1, 2, 3, 4, 5);

if ($arr[0] == 1) {
    say "PASS: array access";
} else {
    say "FAIL: array access";
}

if (scalar(@arr) == 5) {
    say "PASS: array length";
} else {
    say "FAIL: array length";
}

push(@arr, 6);
if (scalar(@arr) == 6) {
    say "PASS: push";
} else {
    say "FAIL: push";
}

my $p = pop(@arr);
if ($p == 6) {
    say "PASS: pop";
} else {
    say "FAIL: pop";
}

my $joined = join("-", @arr);
if ($joined eq "1-2-3-4-5") {
    say "PASS: join";
} else {
    say "FAIL: join got $joined";
}

say "Done!";
