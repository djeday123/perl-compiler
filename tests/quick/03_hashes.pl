# Test hashes

say "Testing hashes...";

my %h = (a => 1, b => 2, c => 3);

if ($h{a} == 1) {
    say "PASS: hash access";
} else {
    say "FAIL: hash access";
}

$h{d} = 4;
if ($h{d} == 4) {
    say "PASS: hash assign";
} else {
    say "FAIL: hash assign";
}

my @k = keys(%h);
if (scalar(@k) == 4) {
    say "PASS: keys";
} else {
    say "FAIL: keys";
}

say "Done!";
