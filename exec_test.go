package mute

import (
	"bufio"
	"bytes"
	"testing"
)

func TestExec(t *testing.T) {
	conf := new(Conf)
	var buf bytes.Buffer
	var bufWriter = bufio.NewWriter(&buf)
	got := Exec("go", []string{"version"}, conf, bufWriter)
	if got != 0 {
		t.Errorf("Exec return val. got: %d want: 0", got)
	}
}

func TestCmdCriteriaReturnDefault(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	conf := new(Conf)
	conf.Default.add(c1)

	got := cmdCriteria("testcommand", conf)
	if !got.equal(&conf.Default) {
		t.Errorf("cmdCriteria should have returned conf default but didn't")
	}
}

func TestCmdCriteriaReturnCommandSpecific(t *testing.T) {
	var crt1, crt2 Criteria
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{1}, []string{})
	crt1 = append(crt1, c1)
	crt2 = append(crt2, c2)

	var commandsCriteria map[string]Criteria
	commandsCriteria = make(map[string]Criteria)
	commandsCriteria["test"] = crt1
	commandsCriteria["testcommand"] = crt2
	commandsCriteria["somethingelse"] = crt1

	conf := new(Conf)
	conf.Commands = commandsCriteria

	got := cmdCriteria("testcommand", conf)
	if !got.equal(&crt2) {
		t.Errorf("cmdCriteria should have returned longest matched cmd but didn't")
	}
}