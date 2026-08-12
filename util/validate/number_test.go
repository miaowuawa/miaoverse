package validate

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestIsDigit(t *testing.T) {
	v := validator.New()
	if err := InitialValidator(v); err != nil {
		t.Fatalf("InitialValidator() error = %v", err)
	}

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "empty", value: "", want: false},
		{name: "ascii digits", value: "13800138000", want: true},
		{name: "leading zero", value: "0123", want: true},
		{name: "space", value: "138 0013", want: false},
		{name: "plus sign", value: "+8613800138000", want: false},
		{name: "letters", value: "abc123", want: false},
		{name: "fullwidth digits", value: "１３８００１３８０００", want: false},
		{name: "arabic-indic digits", value: "١٣٨٠٠١٣٨٠٠٠", want: false},
		{name: "hyphen", value: "138-0013", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Var(tt.value, "isDigit") == nil; got != tt.want {
				t.Fatalf("isDigit(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsPhone(t *testing.T) {
	v := validator.New()
	if err := InitialValidator(v); err != nil {
		t.Fatalf("InitialValidator() error = %v", err)
	}

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "empty", value: "", want: false},
		{name: "valid cn phone", value: "13800138000", want: true},
		{name: "valid min length", value: "12345", want: true},
		{name: "valid max length", value: "123456789012345", want: true},
		{name: "too short", value: "1234", want: false},
		{name: "too long", value: "1234567890123456", want: false},
		{name: "fullwidth digits", value: "１３８００１３８０００", want: false},
		{name: "plus sign", value: "+8613800138000", want: false},
		{name: "letters", value: "13800abc", want: false},
		{name: "space", value: "138 0013", want: false},
		{name: "hyphen", value: "138-0013-8000", want: false},
		{name: "leading zero", value: "013800138000", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Var(tt.value, "isPhone") == nil; got != tt.want {
				t.Fatalf("isPhone(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
