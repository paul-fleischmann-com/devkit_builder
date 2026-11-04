package initcmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffold(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new")
	if err := Scaffold(dir, false); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"base.json", "kits/example.json", "features/git/feature.json"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	st, err := os.Stat(filepath.Join(dir, "features/git/install.sh"))
	if err != nil || st.Mode()&0o100 == 0 {
		t.Errorf("install.sh must be executable: %v %v", st, err)
	}
	if err := Scaffold(dir, false); err == nil {
		t.Error("second scaffold without force must refuse to overwrite")
	}
	if err := Scaffold(dir, true); err != nil {
		t.Errorf("force: %v", err)
	}
}
