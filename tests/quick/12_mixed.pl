# tests/quick/12_mixed.pl
say "Testing mixed arrays...";

# Массив с разными типами
my @mix = (42, "hello", 3.14, [1, 2, 3], {name => "test"});

# Доступ к элементам разных типов
if ($mix[0] == 42) {
    say "PASS: int element";
} else {
    say "FAIL: int element";
}

if ($mix[1] eq "hello") {
    say "PASS: string element";
} else {
    say "FAIL: string element";
}

if ($mix[2] > 3.0) {
    say "PASS: float element";
} else {
    say "FAIL: float element";
}

# Вложенный массив
if ($mix[3][1] == 2) {
    say "PASS: nested array";
} else {
    say "FAIL: nested array";
}

# Вложенный хеш
if ($mix[4]{name} eq "test") {
    say "PASS: nested hash";
} else {
    say "FAIL: nested hash";
}

# Массив хешей (очень частый паттерн)
my @users = (
    {name => "Alice", age => 30},
    {name => "Bob", age => 25},
);

if ($users[0]{name} eq "Alice" && $users[1]{age} == 25) {
    say "PASS: array of hashes";
} else {
    say "FAIL: array of hashes";
}

# Хеш массивов
my %data = (
    numbers => [1, 2, 3],
    letters => ["a", "b", "c"],
);

if ($data{numbers}[0] == 1 && $data{letters}[2] eq "c") {
    say "PASS: hash of arrays";
} else {
    say "FAIL: hash of arrays";
}

say "Done!";