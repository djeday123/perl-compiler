// test_xs/test_xs_compiled.go
package main

import "fmt"

// Runtime (минимальный)
type SV struct {
	iv    int64
	pv    string
	flags uint8
}

const SVf_IOK uint8 = 1
const SVf_POK uint8 = 4

func svInt(i int64) *SV  { return &SV{iv: i, flags: SVf_IOK} }
func svStr(s string) *SV { return &SV{pv: s, flags: SVf_POK} }

func (sv *SV) AsInt() int64     { return sv.iv }
func (sv *SV) AsString() string { return sv.pv }

var _subs = make(map[string]func(...*SV) *SV)

func perl_register_sub(name string, fn func(...*SV) *SV) {
	_subs[name] = fn
}

// === Сгенерированный код ===

func perl_Test_XS_add(args ...*SV) *SV {
	a := args[0].AsInt()
	b := args[1].AsInt()
	var RETVAL *SV
	RETVAL = svInt(int64(a + b))
	return RETVAL
}

func perl_Test_XS_hello(args ...*SV) *SV {
	name := args[0]
	var RETVAL *SV
	str := name.AsString()
	RETVAL = svStr(fmt.Sprintf("Hello, %s!", str))
	return RETVAL
}

func init() {
	perl_register_sub("Test::XS::hello", perl_Test_XS_hello)
	perl_register_sub("Test::XS::add", perl_Test_XS_add)
}

// === Тест ===

func main() {
	// Тест add
	result := perl_Test_XS_add(svInt(3), svInt(5))
	fmt.Printf("add(3, 5) = %d\n", result.AsInt())

	// Тест hello
	result = perl_Test_XS_hello(svStr("World"))
	fmt.Printf("hello(World) = %s\n", result.AsString())
}
