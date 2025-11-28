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

say "Done!";
