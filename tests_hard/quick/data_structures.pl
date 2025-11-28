# tests/quick/data_structures.pl
# Test arrays and hashes

say "Testing data structures...";

# === Arrays ===
say "=== Arrays ===";

# Create and access
my @arr = (1, 2, 3, 4, 5);
if ($arr[0] == 1 && $arr[4] == 5) {
    say "PASS: Array creation and access";
} else {
    say "FAIL: Array creation/access";
}

# Length
if (scalar(@arr) == 5) {
    say "PASS: Array length";
} else {
    say "FAIL: Array length";
}

# Push/pop
push(@arr, 6);
my $popped = pop(@arr);
if ($popped == 6 && scalar(@arr) == 5) {
    say "PASS: Push/pop";
} else {
    say "FAIL: Push/pop";
}

# Shift/unshift
unshift(@arr, 0);
my $shifted = shift(@arr);
if ($shifted == 0 && scalar(@arr) == 5) {
    say "PASS: Shift/unshift";
} else {
    say "FAIL: Shift/unshift";
}

# Join
my $joined = join("-", @arr);
if ($joined eq "1-2-3-4-5") {
    say "PASS: Join";
} else {
    say "FAIL: Join: $joined";
}

# Split
my @split = split(",", "a,b,c");
if ($split[0] eq "a" && $split[1] eq "b" && $split[2] eq "c") {
    say "PASS: Split";
} else {
    say "FAIL: Split";
}

# Reverse
my @rev = reverse(@arr);
if ($rev[0] == 5 && $rev[4] == 1) {
    say "PASS: Reverse";
} else {
    say "FAIL: Reverse";
}

# Negative index
if ($arr[-1] == 5) {
    say "PASS: Negative index";
} else {
    say "FAIL: Negative index";
}

# Range
my @range = (1..5);
if (scalar(@range) == 5 && $range[0] == 1 && $range[4] == 5) {
    say "PASS: Range operator";
} else {
    say "FAIL: Range operator";
}

# === Hashes ===
say "=== Hashes ===";

# Create and access
my %h = (a => 1, b => 2, c => 3);
if ($h{a} == 1 && $h{b} == 2) {
    say "PASS: Hash creation and access";
} else {
    say "FAIL: Hash creation/access";
}

# Assignment
$h{d} = 4;
if ($h{d} == 4) {
    say "PASS: Hash assignment";
} else {
    say "FAIL: Hash assignment";
}

# Keys
my @keys = sort keys %h;
if (scalar(@keys) == 4 && $keys[0] eq "a") {
    say "PASS: Hash keys";
} else {
    say "FAIL: Hash keys";
}

# Values
my @vals = sort values %h;
if (scalar(@vals) == 4 && $vals[0] == 1) {
    say "PASS: Hash values";
} else {
    say "FAIL: Hash values";
}

# Exists
if (exists $h{a} && !exists $h{z}) {
    say "PASS: Hash exists";
} else {
    say "FAIL: Hash exists";
}

# Delete
delete $h{d};
if (!exists $h{d}) {
    say "PASS: Hash delete";
} else {
    say "FAIL: Hash delete";
}

say "Done!";
