package tests

import (
	"os"
	"testing"
)

// ============================================================
// Extended File I/O Tests
// ============================================================

func TestFileIOBasic(t *testing.T) {
	tests := []TestCase{
		{
			Name: "basic write and read",
			Code: `open(my $fh, ">", "basic_write.txt");
print $fh "Hello";
close($fh);

open($fh, "<", "basic_write.txt");
my $content = <$fh>;
close($fh);
say $content;`,
			ExpectedOutput: "Hello",
			CleanupFiles:   []string{"basic_write.txt"},
		},
		{
			Name: "write multiple lines",
			Code: `open(my $fh, ">", "multi_line.txt");
print $fh "Line 1\n";
print $fh "Line 2\n";
print $fh "Line 3\n";
close($fh);

open($fh, "<", "multi_line.txt");
my @lines;
my $line;
$line = <$fh>; chomp($line); push(@lines, $line);
$line = <$fh>; chomp($line); push(@lines, $line);
$line = <$fh>; chomp($line); push(@lines, $line);
close($fh);
say join("|", @lines);`,
			ExpectedOutput: "Line 1|Line 2|Line 3",
			CleanupFiles:   []string{"multi_line.txt"},
		},
		{
			Name: "say to file adds newline",
			Code: `open(my $fh, ">", "say_file.txt");
say $fh "Hello";
close($fh);

open($fh, "<", "say_file.txt");
my $line = <$fh>;
close($fh);
say length($line);  # Should be 6: "Hello\n"`,
			ExpectedOutput: "6",
			CleanupFiles:   []string{"say_file.txt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestFileIOAppend(t *testing.T) {
	tests := []TestCase{
		{
			Name: "append to file",
			Code: `open(my $fh, ">", "append.txt");
print $fh "First\n";
close($fh);

open($fh, ">>", "append.txt");
print $fh "Second\n";
close($fh);

open($fh, "<", "append.txt");
my $l1 = <$fh>; chomp($l1);
my $l2 = <$fh>; chomp($l2);
close($fh);
say "$l1, $l2";`,
			ExpectedOutput: "First, Second",
			CleanupFiles:   []string{"append.txt"},
		},
		{
			Name: "multiple appends",
			Code: `open(my $fh, ">", "multi_append.txt");
print $fh "A\n";
close($fh);

foreach my $letter ("B", "C", "D") {
    open($fh, ">>", "multi_append.txt");
    print $fh "$letter\n";
    close($fh);
}

open($fh, "<", "multi_append.txt");
my @lines;
for (1..4) {
    my $line = <$fh>;
    chomp($line);
    push(@lines, $line);
}
close($fh);
say join("", @lines);`,
			ExpectedOutput: "ABCD",
			CleanupFiles:   []string{"multi_append.txt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestFileIOReadFromExisting(t *testing.T) {
	tests := []TestCase{
		{
			Name: "read single line file",
			Code: `open(my $fh, "<", "single.txt");
my $content = <$fh>;
close($fh);
chomp($content);
say "Content: $content";`,
			ExpectedOutput: "Content: Single line content",
			SetupFiles: map[string]string{
				"single.txt": "Single line content\n",
			},
		},
		{
			Name: "read multi-line file",
			Code: `open(my $fh, "<", "multi.txt");
my @lines;
my $line;
$line = <$fh>;
while (defined $line) {
    chomp($line);
    push(@lines, $line);
    $line = <$fh>;
}
close($fh);
say scalar(@lines);
say $lines[0];
say $lines[2];`,
			ExpectedOutput: "3\nLine A\nLine C",
			SetupFiles: map[string]string{
				"multi.txt": "Line A\nLine B\nLine C\n",
			},
		},
		{
			Name: "read empty file",
			Code: `open(my $fh, "<", "empty.txt");
my $line = <$fh>;
close($fh);
say defined($line) ? "got content" : "empty";`,
			ExpectedOutput: "empty",
			SetupFiles: map[string]string{
				"empty.txt": "",
			},
		},
		{
			Name: "read file without trailing newline",
			Code: `open(my $fh, "<", "no_newline.txt");
my $content = <$fh>;
close($fh);
say "[$content]";`,
			ExpectedOutput: "[Hello World]",
			SetupFiles: map[string]string{
				"no_newline.txt": "Hello World",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestFileIOOpenModes(t *testing.T) {
	tests := []TestCase{
		{
			Name: "3-arg open write",
			Code: `open(my $fh, ">", "three_arg_write.txt");
print $fh "Three arg";
close($fh);

open($fh, "<", "three_arg_write.txt");
my $c = <$fh>;
close($fh);
say $c;`,
			ExpectedOutput: "Three arg",
			CleanupFiles:   []string{"three_arg_write.txt"},
		},
		{
			Name: "3-arg open read",
			Code: `open(my $fh, "<", "read_mode.txt");
my $c = <$fh>;
close($fh);
chomp($c);
say $c;`,
			ExpectedOutput: "Read mode content",
			SetupFiles: map[string]string{
				"read_mode.txt": "Read mode content\n",
			},
		},
		{
			Name: "3-arg open append",
			Code: `open(my $fh, ">", "append_mode.txt");
print $fh "First\n";
close($fh);

open($fh, ">>", "append_mode.txt");
print $fh "Second\n";
close($fh);

open($fh, "<", "append_mode.txt");
my $l1 = <$fh>;
my $l2 = <$fh>;
close($fh);
chomp($l1); chomp($l2);
say "$l1 and $l2";`,
			ExpectedOutput: "First and Second",
			CleanupFiles:   []string{"append_mode.txt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestFileIOErrorHandling(t *testing.T) {
	tests := []TestCase{
		{
			Name: "open nonexistent file returns false",
			Code: `my $result = open(my $fh, "<", "this_file_does_not_exist_xyz.txt");
say $result ? "success" : "failed";`,
			ExpectedOutput: "failed",
		},
		{
			Name: "check open result",
			Code: `if (open(my $fh, "<", "exists.txt")) {
    my $c = <$fh>;
    close($fh);
    chomp($c);
    say "Opened: $c";
} else {
    say "Failed to open";
}`,
			ExpectedOutput: "Opened: File exists",
			SetupFiles: map[string]string{
				"exists.txt": "File exists\n",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestFileIOWithVariables(t *testing.T) {
	tests := []TestCase{
		{
			Name: "filename in variable",
			Code: `my $filename = "var_file.txt";
open(my $fh, ">", $filename);
print $fh "Variable filename";
close($fh);

open($fh, "<", $filename);
my $c = <$fh>;
close($fh);
say $c;`,
			ExpectedOutput: "Variable filename",
			CleanupFiles:   []string{"var_file.txt"},
		},
		{
			Name: "write variable content",
			Code: `my @data = ("Apple", "Banana", "Cherry");
open(my $fh, ">", "data.txt");
foreach my $item (@data) {
    say $fh $item;
}
close($fh);

open($fh, "<", "data.txt");
my @read;
for (1..3) {
    my $line = <$fh>;
    chomp($line);
    push(@read, $line);
}
close($fh);
say join(", ", @read);`,
			ExpectedOutput: "Apple, Banana, Cherry",
			CleanupFiles:   []string{"data.txt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestFileIOMultipleHandles(t *testing.T) {
	tests := []TestCase{
		{
			Name: "two files simultaneously",
			Code: `open(my $fh1, ">", "file1.txt");
open(my $fh2, ">", "file2.txt");
print $fh1 "Content 1";
print $fh2 "Content 2";
close($fh1);
close($fh2);

open($fh1, "<", "file1.txt");
open($fh2, "<", "file2.txt");
my $c1 = <$fh1>;
my $c2 = <$fh2>;
close($fh1);
close($fh2);
say "$c1 and $c2";`,
			ExpectedOutput: "Content 1 and Content 2",
			CleanupFiles:   []string{"file1.txt", "file2.txt"},
		},
		{
			Name: "read from one write to another",
			Code: `open(my $in, "<", "source.txt");
open(my $out, ">", "dest.txt");
my $line;
$line = <$in>;
while (defined $line) {
    $line =~ s/old/new/g;
    print $out $line;
    $line = <$in>;
}
close($in);
close($out);

open(my $check, "<", "dest.txt");
my $result = <$check>;
close($check);
chomp($result);
say $result;`,
			ExpectedOutput: "new data here",
			SetupFiles: map[string]string{
				"source.txt": "old data here\n",
			},
			CleanupFiles: []string{"dest.txt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestFileIOIntegration(t *testing.T) {
	tests := []TestCase{
		{
			Name: "CSV processing",
			Code: `# Write CSV
open(my $fh, ">", "data.csv");
say $fh "name,age,city";
say $fh "Alice,30,NYC";
say $fh "Bob,25,LA";
say $fh "Charlie,35,Chicago";
close($fh);

# Read and process
open($fh, "<", "data.csv");
my $header = <$fh>; # skip header
my @people;
my $line = <$fh>;
while (defined $line) {
    chomp($line);
    my @fields = split(",", $line);
    push(@people, $fields[0]);
    $line = <$fh>;
}
close($fh);

say join(", ", @people);`,
			ExpectedOutput: "Alice, Bob, Charlie",
			CleanupFiles:   []string{"data.csv"},
		},
		{
			Name: "Log file simulation",
			Code: `my @logs = ("INFO: Started", "WARN: Low memory", "ERROR: Failed");
open(my $fh, ">", "app.log");
foreach my $log (@logs) {
    say $fh $log;
}
close($fh);

# Count errors
open($fh, "<", "app.log");
my $errors = 0;
my $line = <$fh>;
while (defined $line) {
    $errors++ if $line =~ /ERROR/;
    $line = <$fh>;
}
close($fh);
say "Errors: $errors";`,
			ExpectedOutput: "Errors: 1",
			CleanupFiles:   []string{"app.log"},
		},
		{
			Name: "Config file reading",
			Code: `open(my $fh, "<", "config.ini");
my %config;
my $line = <$fh>;
while (defined $line) {
    chomp($line);
    if ($line =~ /^(\w+)\s*=\s*(.+)$/) {
        $config{$1} = $2;
    }
    $line = <$fh>;
}
close($fh);

say "host: $config{host}";
say "port: $config{port}";`,
			ExpectedOutput: "host: localhost\nport: 8080",
			SetupFiles: map[string]string{
				"config.ini": "host = localhost\nport = 8080\ndebug = true\n",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// Cleanup function for any leftover test files
func TestCleanup(t *testing.T) {
	files := []string{
		"test_io.txt", "input_test.txt", "append_test.txt", "say_test.txt",
		"three_arg_test.txt", "basic_write.txt", "multi_line.txt", "say_file.txt",
		"append.txt", "multi_append.txt", "single.txt", "multi.txt", "empty.txt",
		"no_newline.txt", "three_arg_write.txt", "read_mode.txt", "append_mode.txt",
		"exists.txt", "var_file.txt", "data.txt", "file1.txt", "file2.txt",
		"source.txt", "dest.txt", "data.csv", "app.log", "config.ini",
		"output.txt", "test_output.txt",
	}

	for _, f := range files {
		os.Remove(f)
	}
}
