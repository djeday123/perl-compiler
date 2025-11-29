my $str = "Name: Alice";
if ($str =~ /Name: (\w+)/) {
    say "matched";
    say $1;
    if ($1 eq "Alice") {
        say "PASS";
    } else {
        say "FAIL: got [$1]";
    }
}
