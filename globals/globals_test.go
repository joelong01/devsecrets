package globals

import (
	"testing"
)

func Test_hidePat(t *testing.T) {

	pat := "glpat-6z456KLG-c2IkdUQMfoT" // a made up gitlab pat
	if HidePat(pat) == pat {
		t.Error("pat wasn't hidden")
	}

	pat = "ghp_EJLDtSLJKFS7234DLSL77989sdfanasdwl0F"
	if HidePat(pat) == pat {
		t.Error("pat wasn't hidden")
	}

	not_pat := "ghp_EJLDtSLJKFS7234DLSL77989sdfanasdwlx"
	if HidePat(not_pat) != not_pat {
		t.Error("not a pat was hidden")
	}

	pat = "gho_EJLDtSLJKFS7234DLSL77989sdfanasdwl0F"
	if HidePat(pat) == pat {
		t.Error("pat that looks like a gh pat not obfuscated")
	}
}
