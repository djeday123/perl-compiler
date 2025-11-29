my $s = "abc";
if ($s =~ /abc/) {
    say "PASS: simple";
} else {
    say "FAIL: simple";
}
my $s2 = "a@b";
if ($s2 =~ /a@b/) {
    say "PASS: with @";
} else {
    say "FAIL: with @";
}