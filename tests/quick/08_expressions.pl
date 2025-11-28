# Test expressions in print/say

say "Testing expressions...";

my $x = 5;
my $y = 3;

if ($x + $y == 8) {
    say "PASS: addition";
} else {
    say "FAIL: addition";
}

if ($x * $y == 15) {
    say "PASS: multiplication";
} else {
    say "FAIL: multiplication";
}

if ($x - $y == 2) {
    say "PASS: subtraction";
} else {
    say "FAIL: subtraction";
}

my $z = $x + $y * 2;
if ($z == 11) {
    say "PASS: precedence";
} else {
    say "FAIL: precedence";
}

sub show_sum {
    my ($a, $b) = @_;
    say $a + $b;
    return $a + $b;
}

if (show_sum(10, 20) == 30) {
    say "PASS: say expr in sub";
} else {
    say "FAIL: say expr in sub";
}

say "Done!";
