// test_xs/Test.xs
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

MODULE = Test::XS  PACKAGE = Test::XS

SV *
hello(name)
    SV *name
    CODE:
        char *str = SvPV_nolen(name);
        RETVAL = newSVpvf("Hello, %s!", str);
    OUTPUT:
        RETVAL

int
add(a, b)
    int a
    int b
    CODE:
        RETVAL = a + b;
    OUTPUT:
        RETVAL