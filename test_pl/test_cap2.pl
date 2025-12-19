my $str = "Name: Alice";
if ($str =~ /Name: (\w+)/) {
    my $cap = $1;
    say "cap: [$cap]";
    say "len: " . length($cap);
    say "Alice len: " . length("Alice");
    if ($cap eq "Alice") {
        say "PASS";
    } else {
        say "FAIL";
    }
}