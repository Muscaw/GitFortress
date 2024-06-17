package system_git

import (
	"os"
	"testing"
)

func Test_isDir_is_directory(t *testing.T) {
	dirName, err := os.MkdirTemp("", "test")
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(dirName)

	isADirectory, err := isDir(dirName)

	if err != nil {
		t.Fatal("could not execute isDir successfully")
	}

	if !isADirectory {
		t.Fatal("created directory is not identified as one")
	}
}

func Test_isDir_is_a_file(t *testing.T) {
	dirName, err := os.MkdirTemp("", "test")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(dirName)
	os.WriteFile(dirName+"/test", []byte("hello world"), 0644)

	isADirectory, err := isDir(dirName + "/test")
	if err != nil {
		t.Fatal("could not execute isDir successfully")
	}

	if isADirectory {
		t.Fatal("created file is not a directory")
	}
}
