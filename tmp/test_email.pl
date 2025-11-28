my $email = "test@example.com";
if ($email =~ /\w+@\w+\.\w+/) {
    say "PASS";
} else {
    say "FAIL";
}
