package mute

import (
	"os"
	"testing"
)

func TestCodesContain(t *testing.T) {
	haystack := []int{1, 2}
	if !codesContain(haystack, 1) {
		t.Errorf("codesContain got 'false' want 'true'")
	}
	if codesContain(haystack, 3) {
		t.Errorf("codesContain got 'true' want 'false'")
	}
}

func TestStdoutPatternsContain(t *testing.T) {
	stdp1 := NewStdoutPattern("hi")
	stdp2 := NewStdoutPattern(".+not[1-9]+so.*simple")
	stdp3 := NewStdoutPattern("I was 3rd")
	stdp4 := NewStdoutPattern(".+not[1-9]+so.*close")

	haystack := []*StdoutPattern{stdp1, stdp2, stdp3}
	if !stdoutPatternsContain(haystack, stdp2) {
		t.Errorf("stdoutPatternsContain got 'false' want 'true'")
	}
	if stdoutPatternsContain(haystack, stdp4) {
		t.Errorf("stdoutPatternsContain got 'true' want 'false'")
	}
}

func TestCriterionEmpty(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{}, []string{""})
	c3 := NewCriterion([]int{}, []string{})

	if c1.IsEmpty() {
		t.Errorf("Criterion with exit codes IsEmpty got 'true' want 'false'")
	}
	if c2.IsEmpty() {
		t.Errorf("Criterion with patterns IsEmpty got 'true' want 'false'")
	}
	if !c3.IsEmpty() {
		t.Errorf("Empty Criterion.IsEmpty 'false' want 'true'")
	}

}

func TestCriterionEqual(t *testing.T) {
	c1 := NewCriterion([]int{0, 1, 2}, []string{})
	c2 := NewCriterion([]int{0, 1, 2}, []string{})

	if !c1.equal(c2) {
		t.Errorf("Criterion.equal got 'false' want 'true'")
	}

	c2 = NewCriterion([]int{0, 2}, []string{})
	if c1.equal(c2) {
		t.Errorf("Criterion.equal unmatched codes got 'true' want 'false'")
	}

	c2 = NewCriterion([]int{0, 1, 2}, []string{"ok"})
	if c1.equal(c2) {
		t.Errorf("Criterion.equal unmatched patterns got 'true' want 'false'")
	}
}

func TestCriteriaEqual(t *testing.T) {
	c1 := NewCriterion([]int{0, 1, 2}, []string{})
	c2 := NewCriterion([]int{0, 1, 2}, []string{})

	criteria1 := new(Criteria)
	criteria1.add(c1, c2)

	criteria2 := new(Criteria)
	criteria2.add(c1, c2)

	if !c1.equal(c2) {
		t.Errorf("Criteria.equal got 'false' want 'true'")
	}

	c3 := NewCriterion([]int{0, 1, 2}, []string{})
	criteria2.add(c3)

	if !c1.equal(c2) {
		t.Errorf("Criteria.equal got 'true' want 'false'")
	}
}

func TestDefaultConf(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	crt1 := new(Criteria)
	crt1.add(c1)

	conf := DefaultConf()
	if !conf.Default.equal(crt1) {
		t.Errorf("DefaultConf().Default didn't match zero exit code")
	}
}

func TestReadConfFileError(t *testing.T) {
	_, err := ReadConfFile("fixtures/no_such_file.toml")
	if err == nil {
		t.Errorf("ReadConfFile should have returned error")
	}
	if _, ok := err.(ConfAccessError); !ok {
		t.Errorf("ReadConfFile should have returned ConfAccessError")
	}
}

func TestReadConfFileSimple(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{1, 2}, []string{"OK"})
	want := new(Conf)
	want.Default.add(c1).add(c2)

	got, err := ReadConfFile("fixtures/simple.toml")
	if err != nil {
		t.Errorf("ReadConfFile had error: %v", err)
	}

	if !want.equal(got) {
		t.Errorf("ReadConfFile simple didn't match want %v got %v", want, got)
	}

	c3 := NewCriterion([]int{1}, []string{})

	extraCodesConf := new(Conf)
	extraCodesConf.Default.add(c1, c3)

	if extraCodesConf.equal(got) {
		t.Errorf("ReadConfFile simple matched extra codes conf")
	}

	missingCodesConf := new(Conf)
	missingCodesConf.Default.add(c1)

	if missingCodesConf.equal(got) {
		t.Errorf("ReadConfFile simple matched missing codes conf")
	}
}

func TestGetCmdConf(t *testing.T) {
	os.Setenv("MUTE_CONFIG", "fixtures/simple.toml")
	defer os.Unsetenv("MUTE_CONFIG")
	want := createSimpleConf()
	defaultConf := DefaultConf()
	got, err := GetCmdConf()
	if err != nil {
		t.Errorf("GetCmdConf simple want no error, got: %v", err)
	}
	if !want.equal(got) {
		t.Errorf("GetCmdConf simple conf want simple %v got %v", want, got)
	}

	os.Setenv("MUTE_CONFIG", "")
	got, err = GetCmdConf()
	if err != nil {
		t.Errorf("GetCmdConf empty env want no error, got: %v", err)
	}
	if !defaultConf.equal(got) {
		t.Errorf("GetCmdConf empty env conf want default conf, got %v", got)
	}
}

func createSimpleConf() *Conf {
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{1, 2}, []string{"OK"})
	conf := new(Conf)
	conf.Default.add(c1).add(c2)
	return conf
}
