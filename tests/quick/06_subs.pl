# Test subroutines

say "Testing subroutines...";

sub add {
    my ($a, $b) = @_;
    return $a + $b;
}

if (add(2, 3) == 5) {
    say "PASS: sub with args";
} else {
    say "FAIL: sub with args";
}

sub calc {
    my ($a, $b) = @_;
    return $a + $b * 2;
}

if (calc(10, 3) == 16) {
    say "PASS: expression in sub";
} else {
    say "FAIL: expression in sub";
}

sub factorial {
    my ($n) = @_;
    if ($n <= 1) {
        return 1;
    }
    return $n * factorial($n - 1);
}

if (factorial(5) == 120) {
    say "PASS: recursive";
} else {
    say "FAIL: recursive";
}

say "Done!";
