package main

import (
	"os/exec"
	"testing"
)

func TestFileExists(t *testing.T) {
	exec.Command("mkdir", "/tmp/filetest")
	a := fileExists("/tmp/filetest")
	if a == true {
		t.Errorf("Fileexists func testing error")
	}
	exec.Command("rm", "/tmp/filetest")
}
