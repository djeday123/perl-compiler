my $str = "Hello World";

if ($str =~ /World/) {
    say "Found World!";
}

if ($str =~ /hello/i) {
    say "Found hello (case insensitive)!";
}

if ($str !~ /Goodbye/) {
    say "Goodbye not found!";
}

my $email = "test@example.com";
if ($email =~ /\w+@\w+\.\w+/) {
    say "Valid email pattern!";
}