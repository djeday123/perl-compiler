# tests/quick/subroutines.pl
# Test subroutines

say "Testing subroutines...";

# Test 1: Simple sub
sub greet {
    return "Hello";
}
if (greet() eq "Hello") {
    say "PASS: Simple sub";
} else {
    say "FAIL: Simple sub";
}

# Test 2: Sub with arguments
sub add {
    my ($a, $b) = @_;
    return $a + $b;
}
if (add(2, 3) == 5) {
    say "PASS: Sub with args";
} else {
    say "FAIL: Sub with args";
}

# Test 3: Sub with array
sub sum {
    my @nums = @_;
    my $total = 0;
    foreach my $n (@nums) {
        $total += $n;
    }
    return $total;
}
if (sum(1, 2, 3, 4, 5) == 15) {
    say "PASS: Sub with array";
} else {
    say "FAIL: Sub with array";
}

# Test 4: Multiple return values
sub minmax {
    my @nums = @_;
    my $min = $nums[0];
    my $max = $nums[0];
    foreach my $n (@nums) {
        $min = $n if $n < $min;
        $max = $n if $n > $max;
    }
    return ($min, $max);
}
my ($min, $max) = minmax(5, 2, 8, 1, 9);
if ($min == 1 && $max == 9) {
    say "PASS: Multiple returns";
} else {
    say "FAIL: Multiple returns";
}

# Test 5: Recursive sub
sub factorial {
    my ($n) = @_;
    return 1 if $n <= 1;
    return $n * factorial($n - 1);
}
if (factorial(5) == 120) {
    say "PASS: Recursive sub";
} else {
    say "FAIL: Recursive sub";
}

# Test 6: Sub modifying array ref
sub double_all {
    my ($arr_ref) = @_;
    foreach my $i (0..$#{$arr_ref}) {
        $arr_ref->[$i] *= 2;
    }
}
my @nums = (1, 2, 3);
double_all(\@nums);
if ($nums[0] == 2 && $nums[1] == 4 && $nums[2] == 6) {
    say "PASS: Sub modifying ref";
} else {
    say "FAIL: Sub modifying ref";
}

# Test 7: Closure-like behavior
sub make_counter {
    my $count = 0;
    return sub {
        $count++;
        return $count;
    };
}

say "Done!";
