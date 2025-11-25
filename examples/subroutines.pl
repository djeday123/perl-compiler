#!/usr/bin/perl
# Subroutine example

sub add {
    my $a = 10;
    my $b = 20;
    return $a + $b;
}

sub factorial {
    my $n = 5;
    my $result = 1;
    my $i = 1;
    while ($i <= $n) {
        $result = $result * $i;
        $i++;
    }
    return $result;
}

my $sum = add();
my $fact = factorial();

print $sum;
print $fact;
