# tests/quick/regex_basic.pl
# Basic regex tests

say "Testing regex...";

# Test 1: Simple match
my $s = "Hello World";
if ($s =~ /World/) {
    say "PASS: Match found";
} else {
    say "FAIL: Match not found";
}

# Test 2: No match
if ($s !~ /Perl/) {
    say "PASS: No match (correct)";
} else {
    say "FAIL: Unexpected match";
}

# Test 3: Case insensitive
$s = "HELLO";
if ($s =~ /hello/i) {
    say "PASS: Case insensitive match";
} else {
    say "FAIL: Case insensitive failed";
}

# Test 4: Capture groups
$s = "Name: Alice";
if ($s =~ /Name: (\w+)/) {
    if ($1 eq "Alice") {
        say "PASS: Capture group works";
    } else {
        say "FAIL: Wrong capture: $1";
    }
} else {
    say "FAIL: Capture match failed";
}

# Test 5: Substitution
$s = "Hello World";
$s =~ s/World/Perl/;
if ($s eq "Hello Perl") {
    say "PASS: Substitution works";
} else {
    say "FAIL: Substitution wrong: $s";
}

# Test 6: Global substitution
$s = "aaa bbb aaa";
$s =~ s/aaa/xxx/g;
if ($s eq "xxx bbb xxx") {
    say "PASS: Global substitution works";
} else {
    say "FAIL: Global substitution wrong: $s";
}

# Test 7: Capture in substitution
$s = "Hello World";
$s =~ s/(\w+) (\w+)/$2 $1/;
if ($s eq "World Hello") {
    say "PASS: Capture in substitution works";
} else {
    say "FAIL: Capture in substitution wrong: $s";
}

say "Done!";
