#!/usr/bin/perl
# Control flow example

my $x = 10;

if ($x > 5) {
    print "x is greater than 5";
} elsif ($x == 5) {
    print "x equals 5";
} else {
    print "x is less than 5";
}

# While loop
my $i = 0;
while ($i < 5) {
    print $i;
    $i++;
}

# For loop
for my $j (1..10) {
    print $j;
}
