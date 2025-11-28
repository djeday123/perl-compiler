# tests/quick/references.pl
# Test references

say "Testing references...";

# === Scalar references ===
say "=== Scalar refs ===";

my $x = 42;
my $ref = \$x;
if ($$ref == 42) {
    say "PASS: Scalar ref deref";
} else {
    say "FAIL: Scalar ref deref";
}

# Modify through ref
$$ref = 100;
if ($x == 100) {
    say "PASS: Modify through ref";
} else {
    say "FAIL: Modify through ref";
}

# === Array references ===
say "=== Array refs ===";

my @arr = (1, 2, 3);
my $arr_ref = \@arr;
if ($arr_ref->[0] == 1 && $arr_ref->[2] == 3) {
    say "PASS: Array ref access";
} else {
    say "FAIL: Array ref access";
}

# Anonymous array
my $anon_arr = [10, 20, 30];
if ($anon_arr->[1] == 20) {
    say "PASS: Anonymous array";
} else {
    say "FAIL: Anonymous array";
}

# Push to array ref
push(@{$arr_ref}, 4);
if (scalar(@arr) == 4) {
    say "PASS: Push to array ref";
} else {
    say "FAIL: Push to array ref";
}

# === Hash references ===
say "=== Hash refs ===";

my %h = (a => 1, b => 2);
my $h_ref = \%h;
if ($h_ref->{a} == 1 && $h_ref->{b} == 2) {
    say "PASS: Hash ref access";
} else {
    say "FAIL: Hash ref access";
}

# Anonymous hash
my $anon_h = {x => 10, y => 20};
if ($anon_h->{x} == 10) {
    say "PASS: Anonymous hash";
} else {
    say "FAIL: Anonymous hash";
}

# Add to hash ref
$h_ref->{c} = 3;
if ($h{c} == 3) {
    say "PASS: Add to hash ref";
} else {
    say "FAIL: Add to hash ref";
}

# === ref() function ===
say "=== ref() function ===";

if (ref($arr_ref) eq "ARRAY") {
    say "PASS: ref() ARRAY";
} else {
    say "FAIL: ref() ARRAY: " . ref($arr_ref);
}

if (ref($h_ref) eq "HASH") {
    say "PASS: ref() HASH";
} else {
    say "FAIL: ref() HASH: " . ref($h_ref);
}

if (ref(\$x) eq "SCALAR") {
    say "PASS: ref() SCALAR";
} else {
    say "FAIL: ref() SCALAR: " . ref(\$x);
}

# === Nested structures ===
say "=== Nested structures ===";

my $data = {
    name => "Alice",
    scores => [90, 85, 95],
    address => {
        city => "NYC",
        zip => "10001"
    }
};

if ($data->{name} eq "Alice") {
    say "PASS: Nested hash access";
} else {
    say "FAIL: Nested hash access";
}

if ($data->{scores}[1] == 85) {
    say "PASS: Array in hash";
} else {
    say "FAIL: Array in hash";
}

if ($data->{address}{city} eq "NYC") {
    say "PASS: Hash in hash";
} else {
    say "FAIL: Hash in hash";
}

say "Done!";
