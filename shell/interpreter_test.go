package shell

import (
	"fmt"
	"testing"
)

func TestInterpret(t *testing.T) {
	_, err := New().Interpret("ls.exe | cat.exe -n > hogehoge")
	fmt.Println(err)
}

func TestMain(t *testing.T) {
	in := []string{`ahahaha ihhihi`, `foo bar`, `"foo bar"`}
	out := makeCmdline(in, in)
	tst := `"ahahaha ihhihi" "foo bar" "\"foo bar\""`
	if out != tst {
		t.Fatalf(`Fail "%s" != "%s"`, out, tst)
	}
}
