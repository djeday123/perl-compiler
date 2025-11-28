# Test basic control flow

say "Testing control flow...";

my $x = 10;

if ($x > 5) {
    say "PASS: if true";
} else {
    say "FAIL: if should be true";
}

if ($x > 20) {
    say "FAIL: should not run";
} else {
    say "PASS: else branch";
}

# While loop
my $i = 0;
my $sum = 0;
while ($i < 5) {
    $sum = $sum + $i;
    $i = $i + 1;
}

if ($sum == 10) {
    say "PASS: while loop";
} else {
    say "FAIL: while loop got $sum";
}

# Foreach
my @arr = (1, 2, 3);
$sum = 0;
foreach my $n (@arr) {
    $sum = $sum + $n;
}

if ($sum == 6) {
    say "PASS: foreach";
} else {
    say "FAIL: foreach got $sum";
}

# Ternary
my $result = $x > 5 ? "big" : "small";
if ($result eq "big") {
    say "PASS: ternary";
} else {
    say "FAIL: ternary";
}

say "Done!";
