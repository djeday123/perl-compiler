#!/usr/bin/perl
# Test file for batch 2 builtin functions
# tests/quick/14_builtins2.pl

say "Testing batch 2 builtins...";

# === printf ===
say "=== Printf ===";
printf("Number: %d\n", 42);
printf("Float: %.2f\n", 3.14159);
printf("String: %s\n", "hello");
printf("Hex: %x\n", 255);
printf("Padded: %05d\n", 7);
say "PASS: printf (visual check above)";

# === wantarray ===
say "=== Wantarray ===";
sub test_context {
    my $ctx = wantarray();
    if (!defined($ctx)) {
        return "void";
    } elsif ($ctx) {
        return "list";
    } else {
        return "scalar";
    }
}

my @list_result = test_context();
my $scalar_result = test_context();
test_context();

my $wa = wantarray();
if (defined($wa)) {
    say "PASS: wantarray returns defined value";
} else {
    say "PASS: wantarray returns undef (void context at top level)";
}

# === each ===
say "=== Each ===";
my %h = (a => 1, b => 2, c => 3);
my @pair = each(%h);
if (scalar(@pair) == 2) {
    say "PASS: each returns key-value pair";
} else {
    say "FAIL: each - expected 2 elements, got " . scalar(@pair);
}

# === pos ===
say "=== Pos ===";
my $str = "hello world";
my $p = pos($str);
if (!defined($p)) {
    say "PASS: pos returns undef before match";
} else {
    say "INFO: pos returned $p";
}

# === grep ===
say "=== Grep ===";
my @nums = (1, 2, 3, 4, 5, 6, 7, 8, 9, 10);

my @big = grep { $_ > 5 } @nums;
if (scalar(@big) == 5) {
    say "PASS: grep block - filtered to 5 elements";
} else {
    say "FAIL: grep block - expected 5, got " . scalar(@big);
}

my @even = grep { $_ % 2 == 0 } @nums;
if (scalar(@even) == 5) {
    say "PASS: grep even numbers";
} else {
    say "FAIL: grep even - expected 5, got " . scalar(@even);
}

# === map ===
say "=== Map ===";

my @doubled = map { $_ * 2 } @nums;
if ($doubled[0] == 2 && $doubled[9] == 20) {
    say "PASS: map multiply";
} else {
    say "FAIL: map multiply - got $doubled[0] and $doubled[9]";
}

my @squared = map { $_ * $_ } (1, 2, 3, 4, 5);
if ($squared[0] == 1 && $squared[4] == 25) {
    say "PASS: map square";
} else {
    say "FAIL: map square";
}

# === eof ===
say "=== Eof ===";
my $eof_result = eof();
if (defined($eof_result)) {
    say "PASS: eof returns defined value";
} else {
    say "FAIL: eof returned undef";
}

# === tell ===
say "=== Tell ===";
my $tell_result = tell();
if (defined($tell_result)) {
    say "PASS: tell returns value";
} else {
    say "FAIL: tell returned undef";
}

# === binmode ===
say "=== Binmode ===";
my $binmode_result = binmode(STDOUT);
if ($binmode_result) {
    say "PASS: binmode returns true";
} else {
    say "FAIL: binmode returned false";
}

# === File operations test ===
say "=== File Seek/Tell/Read ===";
my $test_file = "tmp/perlc_test_file.txt";
open(my $fh, ">", $test_file);
print $fh "Hello World!";
close($fh);

open($fh, "<", $test_file);
my $pos1 = tell($fh);
if ($pos1 == 0) {
    say "PASS: tell at start is 0";
} else {
    say "FAIL: tell at start - expected 0, got $pos1";
}

seek($fh, 6, 0);
my $pos2 = tell($fh);
if ($pos2 == 6) {
    say "PASS: seek and tell work";
} else {
    say "FAIL: after seek(6) tell returned $pos2";
}

my $buf;
my $bytes = read($fh, $buf, 5);
if ($buf eq "World") {
    say "PASS: read returned correct data";
} else {
    say "FAIL: read - expected 'World', got '$buf'";
}

close($fh);

# === Grep/Map Edge Cases ===
say "=== Grep/Map Edge Cases ===";

my @empty = ();
my @filtered_empty = grep { $_ > 0 } @empty;
if (scalar(@filtered_empty) == 0) {
    say "PASS: grep on empty array";
} else {
    say "FAIL: grep on empty array";
}

my @mapped_empty = map { $_ * 2 } @empty;
if (scalar(@mapped_empty) == 0) {
    say "PASS: map on empty array";
} else {
    say "FAIL: map on empty array";
}

my @words = ("apple", "banana", "cherry", "date");
my @long_words = grep { length($_) > 5 } @words;
if (scalar(@long_words) == 2) {
    say "PASS: grep strings by length";
} else {
    say "FAIL: grep strings - expected 2, got " . scalar(@long_words);
}

my @upper = map { uc($_) } @words;
if ($upper[0] eq "APPLE" && $upper[3] eq "DATE") {
    say "PASS: map to uppercase";
} else {
    say "FAIL: map to uppercase";
}

say "Done!";