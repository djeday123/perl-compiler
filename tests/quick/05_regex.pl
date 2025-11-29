		   
# Test regex

say "Testing regex...";

my $s = "Hello World";

if ($s =~ /World/) {
    say "PASS: match";
} else {
    say "FAIL: match";
}

if ($s !~ /Perl/) {
    say "PASS: no match";
} else {
    say "FAIL: no match";
}

$s =~ s/World/Perl/;
if ($s eq "Hello Perl") {
    say "PASS: substitution";
} else {
    say "FAIL: subst got $s";
}

my $t = "aaa";
$t =~ s/a/b/g;
if ($t eq "bbb") {
    say "PASS: global subst";
} else {
    say "FAIL: global subst got $t";
}

# Complex pattern - email
my $email = 'test@example.com';
if ($email =~ /\w+\@\w+\.\w+/) {
    say "PASS: complex pattern";
} else {
    say "FAIL: complex pattern";
}

# Case insensitive
my $text = "HELLO";
if ($text =~ /hello/i) {
    say "PASS: case insensitive";
} else {
    say "FAIL: case insensitive";
}

# Capture group
my $str = "Name: Alice";
if ($str =~ /Name: (\w+)/) {
    if ($1 eq "Alice") {
        say "PASS: capture group";
    } else {
        say "FAIL: capture wrong: $1";
    }
} else {
    say "FAIL: capture match";
}

# Substitution with capture
$str = "Hello World";
$str =~ s/(\w+) (\w+)/$2 $1/;
if ($str eq "World Hello") {
    say "PASS: subst with capture";
} else {
    say "FAIL: subst capture: $str";
}


say "Done!";
# === Nested data structures ===
