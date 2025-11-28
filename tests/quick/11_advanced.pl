# Advanced tests - features not in basic tests
say "Testing advanced features...";

# === Until loop ===
say "=== Until loop ===";
my $i = 0;
until ($i >= 3) {
    $i = $i + 1;
}
if ($i == 3) {
    say "PASS: until loop";
} else {
    say "FAIL: until loop - got $i";
}

# === Elsif ===
say "=== Elsif ===";
my $x = 15;
if ($x > 20) {
    say "FAIL: elsif";
} elsif ($x > 10) {
    say "PASS: elsif branch";
} else {
    say "FAIL: elsif";
}

# === Unless ===
say "=== Unless ===";
unless ($x > 100) {
    say "PASS: unless";
}

# === Statement modifiers ===
say "=== Statement modifiers ===";
say "PASS: if modifier" if $x > 10;
say "PASS: unless modifier" unless $x > 100;

# === For C-style loop ===
say "=== For C-style ===";
my $sum = 0;
for (my $j = 0; $j < 5; $j++) {
    $sum = $sum + $j;
}
if ($sum == 10) {
    say "PASS: for C-style";
} else {
    say "FAIL: for C-style - got $sum";
}

# === Foreach with range ===
say "=== Foreach range ===";
$sum = 0;
foreach my $n (1..5) {
    $sum = $sum + $n;
}
if ($sum == 15) {
    say "PASS: foreach range";
} else {
    say "FAIL: foreach range - got $sum";
}

# === Loop control: last ===
say "=== Last ===";
$sum = 0;
foreach my $n (1..10) {
    if ($n > 5) {
        last;
    }
    $sum = $sum + $n;
}
if ($sum == 15) {
    say "PASS: last";
} else {
    say "FAIL: last - got $sum";
}

# === Loop control: next ===
say "=== Next ===";
$sum = 0;
foreach my $n (1..10) {
    if ($n % 2 == 0) {
        next;
    }
    $sum = $sum + $n;
}
if ($sum == 25) {
    say "PASS: next";
} else {
    say "FAIL: next - got $sum";
}

# === Logical and returns value ===
say "=== Logical operators ===";
my $result = 1 && 2;
if ($result == 2) {
    say "PASS: logical and value";
} else {
    say "FAIL: logical and value - got $result";
}

# === Logical or returns value ===
$result = 0 || 42;
if ($result == 42) {
    say "PASS: logical or value";
} else {
    say "FAIL: logical or value - got $result";
}

# === Defined-or ===
say "=== Defined-or ===";
my $undef;
$result = $undef // "default";
if ($result eq "default") {
    say "PASS: defined-or";
} else {
    say "FAIL: defined-or";
}

# === Shift/unshift ===
say "=== Shift/unshift ===";
my @arr = (1, 2, 3);
unshift(@arr, 0);
my $shifted = shift(@arr);
if ($shifted == 0 && scalar(@arr) == 3) {
    say "PASS: shift/unshift";
} else {
    say "FAIL: shift/unshift";
}

# === Negative array index ===
say "=== Negative index ===";
@arr = (1, 2, 3, 4, 5);
if ($arr[-1] == 5 && $arr[-2] == 4) {
    say "PASS: negative index";
} else {
    say "FAIL: negative index";
}

# === Split ===
say "=== Split ===";
my @parts = split(",", "a,b,c");
if ($parts[0] eq "a" && $parts[1] eq "b" && $parts[2] eq "c") {
    say "PASS: split";
} else {
    say "FAIL: split";
}

# === Reverse ===
say "=== Reverse ===";
@arr = (1, 2, 3);
my @rev = reverse(@arr);
if ($rev[0] == 3 && $rev[2] == 1) {
    say "PASS: reverse";
} else {
    say "FAIL: reverse";
}

# === Sort ===
say "=== Sort ===";
@arr = (3, 1, 2);
my @sorted = sort(@arr);
if ($sorted[0] == 1 && $sorted[2] == 3) {
    say "PASS: sort";
} else {
    say "FAIL: sort";
}

# === Hash values ===
say "=== Hash values ===";
my %h = (a => 1, b => 2);
my @vals = values(%h);
if (scalar(@vals) == 2) {
    say "PASS: values";
} else {
    say "FAIL: values";
}

# === Exists ===
say "=== Exists ===";
if (exists $h{a}) {
    say "PASS: exists true";
} else {
    say "FAIL: exists true";
}
if (!exists $h{z}) {
    say "PASS: exists false";
} else {
    say "FAIL: exists false";
}

# === Delete ===
say "=== Delete ===";
$h{temp} = 99;
delete $h{temp};
if (!exists $h{temp}) {
    say "PASS: delete";
} else {
    say "FAIL: delete";
}

# === Scalar reference ===
say "=== Scalar refs ===";
my $val = 42;
my $ref = \$val;
if ($$ref == 42) {
    say "PASS: scalar ref deref";
} else {
    say "FAIL: scalar ref deref";
}

$$ref = 100;
if ($val == 100) {
    say "PASS: modify through scalar ref";
} else {
    say "FAIL: modify through scalar ref";
}

# === Array ref with backslash ===
say "=== Array ref backslash ===";
@arr = (10, 20, 30);
my $arr_ref = \@arr;
if ($arr_ref->[1] == 20) {
    say "PASS: array ref backslash";
} else {
    say "FAIL: array ref backslash";
}

# === Hash ref with backslash ===
say "=== Hash ref backslash ===";
%h = (x => 5, y => 10);
my $h_ref = \%h;
if ($h_ref->{x} == 5) {
    say "PASS: hash ref backslash";
} else {
    say "FAIL: hash ref backslash";
}

# === ref() for SCALAR ===
say "=== ref SCALAR ===";
if (ref(\$val) eq "SCALAR") {
    say "PASS: ref SCALAR";
} else {
    say "FAIL: ref SCALAR - got " . ref(\$val);
}

# === Nested data structures ===
say "=== Nested structures ===";
my $data = {
    name => "Test",
    nums => [1, 2, 3],
    info => { id => 42 }
};
if ($data->{name} eq "Test") {
    say "PASS: nested hash";
} else {
    say "FAIL: nested hash";
}
if ($data->{nums}[1] == 2) {
    say "PASS: array in hash";
} else {
    say "FAIL: array in hash";
}
if ($data->{info}{id} == 42) {
    say "PASS: hash in hash";
} else {
    say "FAIL: hash in hash";
}

# === Regex capture groups ===
say "=== Regex capture ===";
my $s = "Name: Alice";
if ($s =~ /Name: (\w+)/) {
    if ($1 eq "Alice") {
        say "PASS: capture group";
    } else {
        say "FAIL: capture group - got $1";
    }
} else {
    say "FAIL: capture match";
}

# === Case insensitive regex ===
say "=== Regex /i ===";
$s = "HELLO";
if ($s =~ /hello/i) {
    say "PASS: case insensitive";
} else {
    say "FAIL: case insensitive";
}

# === Capture in substitution ===
say "=== Subst capture ===";
$s = "Hello World";
$s =~ s/(\w+) (\w+)/$2 $1/;
if ($s eq "World Hello") {
    say "PASS: subst capture";
} else {
    say "FAIL: subst capture - got $s";
}

# === Chomp ===
say "=== Chomp ===";
$s = "test\n";
chomp($s);
if ($s eq "test") {
    say "PASS: chomp";
} else {
    say "FAIL: chomp";
}

# === Multiple return values ===
say "=== Multiple returns ===";
sub minmax {
    my @nums = @_;
    my $min = $nums[0];
    my $max = $nums[0];
    foreach my $n (@nums) {
        if ($n < $min) { $min = $n; }
        if ($n > $max) { $max = $n; }
    }
    return ($min, $max);
}
my ($min, $max) = minmax(5, 2, 8, 1, 9);
if ($min == 1 && $max == 9) {
    say "PASS: multiple return";
} else {
    say "FAIL: multiple return";
}

# === Increment/decrement ===
say "=== Inc/Dec ===";
my $n = 5;
$n++;
if ($n == 6) {
    say "PASS: post-increment";
} else {
    say "FAIL: post-increment";
}
$n--;
if ($n == 5) {
    say "PASS: post-decrement";
} else {
    say "FAIL: post-decrement";
}

# === Compound assignment ===
say "=== Compound assignment ===";
$n = 10;
$n += 5;
if ($n == 15) {
    say "PASS: +=";
} else {
    say "FAIL: +=";
}
$n -= 3;
if ($n == 12) {
    say "PASS: -=";
} else {
    say "FAIL: -=";
}
$n *= 2;
if ($n == 24) {
    say "PASS: *=";
} else {
    say "FAIL: *=";
}
my $str = "Hello";
$str .= " World";
if ($str eq "Hello World") {
    say "PASS: .=";
} else {
    say "FAIL: .=";
}


# === Defined ===
say "=== Defined ===";
my $undef_var;
if (!defined($undef_var)) {
    say "PASS: undef detected";
} else {
    say "FAIL: undef not detected";
}
$undef_var = 42;
if (defined($undef_var)) {
    say "PASS: defined detected";
} else {
    say "FAIL: defined not detected";
}


say "Done!";