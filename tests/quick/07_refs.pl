# Test references (anonymous only)

say "Testing references...";

my $arr = [1, 2, 3];

if ($arr->[0] == 1) {
    say "PASS: array ref access";
} else {
    say "FAIL: array ref";
}

my $h = {a => 10, b => 20};

if ($h->{a} == 10) {
    say "PASS: hash ref access";
} else {
    say "FAIL: hash ref";
}

if (ref($arr) eq "ARRAY") {
    say "PASS: ref type ARRAY";
} else {
    say "FAIL: ref type";
}

if (ref($h) eq "HASH") {
    say "PASS: ref type HASH";
} else {
    say "FAIL: ref type";
}

say "Done!";
