# OOP tests
say "Testing OOP...";

# Test 1: Basic bless
my $obj = {};
bless($obj, "MyClass");
my $type = ref($obj);
if ($type eq "MyClass") {
    say "PASS: basic bless";
} else {
    say "FAIL: basic bless - got $type";
}

# Test 2: Constructor pattern
sub Point::new {
    my $class = shift;
    my $x = shift;
    my $y = shift;
    my $self = {};
    $self->{"x"} = $x;
    $self->{"y"} = $y;
    bless($self, $class);
    return $self;
}

sub Point::get_x {
    my $self = shift;
    return $self->{"x"};
}

sub Point::get_y {
    my $self = shift;
    return $self->{"y"};
}

my $p = Point->new(10, 20);
my $ptype = ref($p);
if ($ptype eq "Point") {
    say "PASS: constructor";
} else {
    say "FAIL: constructor - got $ptype";
}

# Test 3: Method calls
my $px = $p->get_x();
if ($px == 10) {
    say "PASS: method get_x";
} else {
    say "FAIL: method get_x - got $px";
}

my $py = $p->get_y();
if ($py == 20) {
    say "PASS: method get_y";
} else {
    say "FAIL: method get_y - got $py";
}

# Test 4: Inheritance
sub Shape::area {
    return 0;
}

set_isa("Rectangle", "Shape");

sub Rectangle::new {
    my $class = shift;
    my $w = shift;
    my $h = shift;
    my $self = {};
    $self->{"width"} = $w;
    $self->{"height"} = $h;
    bless($self, $class);
    return $self;
}

sub Rectangle::area {
    my $self = shift;
    return $self->{"width"} * $self->{"height"};
}

my $rect = Rectangle->new(5, 4);
my $area = $rect->area();
if ($area == 20) {
    say "PASS: inheritance area";
} else {
    say "FAIL: inheritance area - got $area";
}

say "Done!";