package UserProfile

import "testing"

func TestFormatMaskedPhone(t *testing.T) {
	cases := []struct {
		name   string
		phone  string
		region uint16
		want   string
	}{
		{"中国大陆手机号", "13800138000", 86, "+86 138****8000"},
		{"无区号", "13800138000", 0, "138****8000"},
		{"短号码整体打码", "1234567", 86, "+86 *******"},
		{"空号码", "", 86, ""},
		{"前后空白", "  13800138000  ", 1, "+1 138****8000"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FormatMaskedPhone(c.phone, c.region); got != c.want {
				t.Fatalf("FormatMaskedPhone(%q, %d) = %q, want %q", c.phone, c.region, got, c.want)
			}
		})
	}
}

func TestMaskPhoneNeverLeaksMiddleDigits(t *testing.T) {
	got := MaskPhone("13812345678")
	if got != "138****5678" {
		t.Fatalf("MaskPhone = %q, want %q", got, "138****5678")
	}
	if len(got) != len("13812345678") {
		t.Fatalf("MaskPhone length = %d, want same as input", len(got))
	}
}
