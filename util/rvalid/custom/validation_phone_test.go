package custom

import (
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
)

var onceTestRegisterPhoneValidation sync.Once

func TestRegisterPhoneValidation(t *testing.T) {
	initValidator()
	onceTestRegisterPhoneValidation.Do(func() {
		if err := RegisterPhoneValidation(validate, zhTranslator); err != nil {
			t.Errorf("unexpected register: %s", err)
		}
	})
}

func TestValidateCnPhone(t *testing.T) {
	TestRegisterPhoneValidation(t)
	type testStruct struct {
		Phone string `v:"cnphone"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"valid_13x", "13800138000", false},
		{"valid_14x", "14700000000", false},
		{"valid_15x", "15012345678", false},
		{"valid_16x", "16600000000", false},
		{"valid_17x", "17700000000", false},
		{"valid_18x", "18812345678", false},
		{"valid_19x", "19900000000", false},
		{"empty", "", true},
		{"too_short", "1380013800", true},
		{"too_long", "138001380001", true},
		{"start_with_0", "03800138000", true},
		{"start_with_2", "23800138000", true},
		{"second_digit_0", "10800138000", true},
		{"second_digit_1", "11800138000", true},
		{"second_digit_2", "12800138000", true},
		{"has_letter", "1380013800a", true},
		{"has_space", "138 0013 8000", true},
		{"has_dash", "138-0013-8000", true},
		{"has_plus", "+8613800138000", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Phone: tt.val}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("cnphone(%q) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestCnPhoneTranslation(t *testing.T) {
	TestRegisterPhoneValidation(t)

	type testStruct struct {
		Phone string `v:"cnphone"`
	}

	s := testStruct{Phone: "abc"}
	err := validate.Struct(s)
	if err == nil {
		t.Fatal("expected error")
	}
	errs := err.(validator.ValidationErrors)
	msg := errs[0].Translate(zhTranslator)
	if !strings.Contains(msg, "手机号") {
		t.Errorf("unexpected translation: %s", msg)
	}
}
