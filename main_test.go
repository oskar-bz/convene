package main

import "testing"
import "time"

func TestSplitFileFromPath(t *testing.T) {
	path := "C:\\users\\ossel\\downloads\\moin.abc"
	got_path, got_file := SplitFileFromPath(path)
	if got_path != "C:\\users\\ossel\\downloads\\" {
		t.Errorf("SplitFileFromPath returned wrong path, got: %s", got_path)
	}
	if got_file != "moin.abc" {
		t.Errorf("SplitFileFromPath returned wrong file, got %s", got_file)
	}

	path = "servus/moin/servus.txt"
	got_path, got_file = SplitFileFromPath(path)
	if got_path != "servus/moin/" {
		t.Errorf("SplitFileFromPath returned wrong path, got: %s", got_path)
	}
	if got_file != "servus.txt" {
		t.Errorf("SplitFileFromPath returned wrong file, got %s", got_file)
	}
}

func TestMovePath(t *testing.T) {
	src := "this/is/my/file.exe"
	dst := "new/directory/"
	base := "this/is/"
	got := MovePath(src, dst, base)
	if got != "new/directory/my/file.exe" {
		t.Errorf("MovePath, got %s", got)
	}
}

func TestTimeToString(t *testing.T) {
	t.Errorf("%s", time.Now().Local().Format(time.RFC1123))
}
