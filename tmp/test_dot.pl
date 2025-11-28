my $s = "a.b";
if ($s =~ /a\.b/) {
    say "PASS: escaped dot";
} else {
    say "FAIL: escaped dot";
}
if ($s =~ /a.b/) {
    say "PASS: any char";
} else {
    say "FAIL: any char";
}