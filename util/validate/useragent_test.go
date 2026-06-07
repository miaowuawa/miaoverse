package validate

import "testing"

func TestParseUA(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{
			name: "pc",
			ua:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/125.0.0.0 Safari/537.36",
			want: UATypePC,
		},
		{
			name: "wap",
			ua:   "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1",
			want: UATypeWAP,
		},
		{
			name: "bot",
			ua:   "Googlebot/2.1 (+http://www.google.com/bot.html)",
			want: UATypeBot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseUA(tt.ua); got != tt.want {
				t.Fatalf("ParseUA() = %q, want %q", got, tt.want)
			}
		})
	}
}
