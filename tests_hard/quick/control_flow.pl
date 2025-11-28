# tests/quick/control_flow.pl
# Test control flow statements

say "Testing control flow...";

# === If/else ===
say "=== If/else ===";

my $x = 10;
if ($x > 5) {
    say "PASS: if true";
}

if ($x > 20) {
    say "FAIL: should not print";
} else {
    say "PASS: else branch";
}

if ($x > 20) {
    say "FAIL: should not print";
} elsif ($x > 5) {
    say "PASS: elsif branch";
} else {
    say "FAIL: should not print";
}

# === Unless ===
say "=== Unless ===";

unless ($x > 20) {
    say "PASS: unless works";
}

# === Statement modifiers ===
say "=== Statement modifiers ===";

say "PASS: if modifier" if $x > 5;
say "PASS: unless modifier" unless $x > 20;

# === Loops ===
say "=== Loops ===";

# While
my $i = 0;
my $sum = 0;
while ($i < 5) {
    $sum += $i;
    $i++;
}
if ($sum == 10) {
    say "PASS: while loop";
} else {
    say "FAIL: while loop: $sum";
}

# Until
$i = 0;
$sum = 0;
until ($i >= 5) {
    $sum += $i;
    $i++;
}
if ($sum == 10) {
    say "PASS: until loop";
} else {
    say "FAIL: until loop";
}

# For C-style
$sum = 0;
for (my $j = 0; $j < 5; $j++) {
    $sum += $j;
}
if ($sum == 10) {
    say "PASS: for loop";
} else {
    say "FAIL: for loop";
}

# Foreach
my @arr = (1, 2, 3, 4, 5);
$sum = 0;
foreach my $n (@arr) {
    $sum += $n;
}
if ($sum == 15) {
    say "PASS: foreach loop";
} else {
    say "FAIL: foreach loop";
}

# Foreach with range
$sum = 0;
foreach my $n (1..5) {
    $sum += $n;
}
if ($sum == 15) {
    say "PASS: foreach range";
} else {
    say "FAIL: foreach range";
}

# Last
say "=== Loop control ===";
$sum = 0;
foreach my $n (1..10) {
    last if $n > 5;
    $sum += $n;
}
if ($sum == 15) {
    say "PASS: last";
} else {
    say "FAIL: last";
}

# Next
$sum = 0;
foreach my $n (1..10) {
    next if $n % 2 == 0;
    $sum += $n;
}
if ($sum == 25) {  # 1+3+5+7+9
    say "PASS: next";
} else {
    say "FAIL: next: $sum";
}

# === Ternary ===
say "=== Ternary ===";
my $result = $x > 5 ? "big" : "small";
if ($result eq "big") {
    say "PASS: ternary";
} else {
    say "FAIL: ternary";
}

# === Logical operators ===
say "=== Logical operators ===";

# And
$result = 1 && 2;
if ($result == 2) {
    say "PASS: logical and";
} else {
    say "FAIL: logical and";
}

# Or
$result = 0 || "default";
if ($result eq "default") {
    say "PASS: logical or";
} else {
    say "FAIL: logical or";
}

# Defined-or
my $undef;
$result = $undef // "fallback";
if ($result eq "fallback") {
    say "PASS: defined-or";
} else {
    say "FAIL: defined-or";
}

say "Done!";
