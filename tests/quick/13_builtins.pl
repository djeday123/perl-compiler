#!/usr/bin/perl
# Test file for new builtin functions
# tests/quick/13_builtins.pl

say "Testing new builtin functions...";

# === index/rindex ===
say "=== Index/Rindex ===";
my $text = "Hello World Hello";
if (index($text, "World") == 6) {
    say "PASS: index";
} else {
    say "FAIL: index - got " . index($text, "World");
}
if (rindex($text, "Hello") == 12) {
    say "PASS: rindex";
} else {
    say "FAIL: rindex - got " . rindex($text, "Hello");
}
if (index($text, "xyz") == -1) {
    say "PASS: index not found";
} else {
    say "FAIL: index not found";
}
# index with position
if (index($text, "Hello", 1) == 12) {
    say "PASS: index with position";
} else {
    say "FAIL: index with position - got " . index($text, "Hello", 1);
}

# === lcfirst/ucfirst ===
say "=== Lcfirst/Ucfirst ===";
if (lcfirst("HELLO") eq "hELLO") {
    say "PASS: lcfirst";
} else {
    say "FAIL: lcfirst - got " . lcfirst("HELLO");
}
if (ucfirst("hello") eq "Hello") {
    say "PASS: ucfirst";
} else {
    say "FAIL: ucfirst - got " . ucfirst("hello");
}
if (ucfirst("hello world") eq "Hello world") {
    say "PASS: ucfirst sentence";
} else {
    say "FAIL: ucfirst sentence";
}

# === chop ===
say "=== Chop ===";
my $s = "hello";
my $last = chop($s);
if ($s eq "hell" && $last eq "o") {
    say "PASS: chop";
} else {
    say "FAIL: chop - s=$s, last=$last";
}

# === sprintf ===
say "=== Sprintf ===";
my $formatted = sprintf("Name: %s, Age: %d", "Alice", 30);
if ($formatted eq "Name: Alice, Age: 30") {
    say "PASS: sprintf basic";
} else {
    say "FAIL: sprintf basic - got $formatted";
}

my $padded = sprintf("%05d", 42);
if ($padded eq "00042") {
    say "PASS: sprintf padding";
} else {
    say "FAIL: sprintf padding - got $padded";
}

my $floatfmt = sprintf("%.2f", 3.14159);
if ($floatfmt eq "3.14") {
    say "PASS: sprintf float";
} else {
    say "FAIL: sprintf float - got $floatfmt";
}

# === quotemeta ===
say "=== Quotemeta ===";
my $meta = quotemeta("a.b*c?");
if ($meta eq "a\\.b\\*c\\?") {
    say "PASS: quotemeta";
} else {
    say "FAIL: quotemeta - got $meta";
}

# === hex ===
say "=== Hex ===";
if (hex("ff") == 255) {
    say "PASS: hex lowercase";
} else {
    say "FAIL: hex lowercase - got " . hex("ff");
}
if (hex("FF") == 255) {
    say "PASS: hex uppercase";
} else {
    say "FAIL: hex uppercase";
}
if (hex("0xFF") == 255) {
    say "PASS: hex with prefix";
} else {
    say "FAIL: hex with prefix";
}
if (hex("10") == 16) {
    say "PASS: hex 10";
} else {
    say "FAIL: hex 10 - got " . hex("10");
}

# === oct ===
say "=== Oct ===";
if (oct("77") == 63) {
    say "PASS: oct";
} else {
    say "FAIL: oct - got " . oct("77");
}
if (oct("0777") == 511) {
    say "PASS: oct with prefix";
} else {
    say "FAIL: oct with prefix - got " . oct("0777");
}
if (oct("0b1111") == 15) {
    say "PASS: oct binary";
} else {
    say "FAIL: oct binary - got " . oct("0b1111");
}
if (oct("0x10") == 16) {
    say "PASS: oct hex";
} else {
    say "FAIL: oct hex - got " . oct("0x10");
}

# === fc ===
say "=== Fc ===";
if (fc("HeLLo") eq "hello") {
    say "PASS: fc";
} else {
    say "FAIL: fc - got " . fc("HeLLo");
}

# === pack/unpack (basic) ===
say "=== Pack/Unpack ===";
my $packed = pack("A3", "abc");
if ($packed eq "abc") {
    say "PASS: pack A3";
} else {
    say "FAIL: pack A3";
}

my $packed_num = pack("C", 65);
if ($packed_num eq "A") {
    say "PASS: pack C";
} else {
    say "FAIL: pack C - got ord=" . ord($packed_num);
}

my $packed_multi = pack("A3A3", "abc", "def");
if ($packed_multi eq "abcdef") {
    say "PASS: pack multiple";
} else {
    say "FAIL: pack multiple - got $packed_multi";
}

my @unpacked = unpack("A3", "abc");
if ($unpacked[0] eq "abc") {
    say "PASS: unpack A3";
} else {
    say "FAIL: unpack A3 - got $unpacked[0]";
}

my @unpacked_c = unpack("C", "A");
if ($unpacked_c[0] == 65) {
    say "PASS: unpack C";
} else {
    say "FAIL: unpack C - got $unpacked_c[0]";
}


# my @unpacked = unpack("A3", "abc");
# if ($unpacked[0] eq "abc") {
#     say "PASS: unpack A3";
# } else {
#     my $got = $unpacked[0];
#     say "FAIL: unpack A3 - got $got";
# }

# my @unpacked_c = unpack("C", "A");
# if ($unpacked_c[0] == 65) {
#     say "PASS: unpack C";
# } else {
#     my $got = $unpacked_c[0];
#     say "FAIL: unpack C - got $got";
# }


say "Done!";