#!/usr/bin/perl
use strict;

sub factorial {
    my $n = shift;
    if ($n <= 1) {
        return 1;
    }
    return $n * factorial($n - 1);
}

say "Factorial calculator";
for (my $i = 1; $i <= 10; $i++) {
    my $f = factorial($i);
    say "$i! = $f";
}