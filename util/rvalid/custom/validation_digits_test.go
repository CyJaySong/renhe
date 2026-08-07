package custom

import (
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var onceTestRegisterDigitsValidation sync.Once

func TestRegisterDigitsValidation(t *testing.T) {
	initValidator()
	onceTestRegisterDigitsValidation.Do(func() {
		RegisterDecimalType(validate)
		if err := RegisterDigitsValidation(validate, zhTranslator); err != nil {
			t.Errorf("unexpected register: %s", err)
		}
	})
}

func TestFloatDigits(t *testing.T) {
	TestRegisterDigitsValidation(t)

	tests := []struct {
		name    string
		val     float64
		wantInt int
		wantFrc int
	}{
		{"zero", 0, 0, 0},
		{"1", 1, 1, 0},
		{"123", 123, 3, 0},
		{"-123", -123, 3, 0},
		{"0.12", 0.12, 0, 2},
		{"123.45", 123.45, 3, 2},
		{"-123.45", -123.45, 3, 2},
		{"0.001", 0.001, 0, 3},
		{"1000", 1000, 4, 0},
		{"99999999.99", 99999999.99, 8, 2},
		{"99.0", 99.0, 2, 0},
		{"0.1", 0.1, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intLen, fracLen := floatDigits(tt.val)
			if intLen != tt.wantInt || fracLen != tt.wantFrc {
				t.Errorf("floatDigits(%v) = (%d, %d), want (%d, %d)",
					tt.val, intLen, fracLen, tt.wantInt, tt.wantFrc)
			}
		})
	}
}

func TestParseDigitsParam(t *testing.T) {
	tests := []struct {
		param   string
		wantInt int
		wantFrc int
	}{
		{"10.2", 10, 2},
		{".2", -1, 2},
		{"10.", 10, -1},
		{"10", 10, -1},
		{"0.2", 0, 2},
	}
	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			mi, mf := parseDigitsParam(tt.param)
			if mi != tt.wantInt || mf != tt.wantFrc {
				t.Errorf("parseDigitsParam(%q) = (%d, %d), want (%d, %d)",
					tt.param, mi, mf, tt.wantInt, tt.wantFrc)
			}
		})
	}
}

func TestValidateDigitsTag(t *testing.T) {
	TestRegisterDigitsValidation(t)

	type testStruct struct {
		Price decimal.Decimal `v:"digits=10.2"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"ok", "123.45", false},
		{"ok_int_only", "1234567890", false},
		{"ok_zero", "0", false},
		{"ok_frac_only", "0.12", false},
		{"fail_int_too_long", "12345678901.12", true},
		{"fail_frac_too_long", "123.456", true},
		{"fail_both", "12345678901.456", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Price: decimal.RequireFromString(tt.val)}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%s) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDigitsString(t *testing.T) {
	TestRegisterDigitsValidation(t)

	type testStruct struct {
		Amount string `v:"digits=2.2"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"ok", "12.34", false},
		{"ok_negative", "-12.34", false},
		{"ok_zero_int", "0.12", false},
		{"fail_int_too_long", "123.45", true},
		{"fail_frac_too_long", "12.345", true},
		{"fail_not_number", "not-a-number", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Amount: tt.val}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%q) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDigitsFracOnly(t *testing.T) {
	TestRegisterDigitsValidation(t)

	type testStruct struct {
		Rate decimal.Decimal `v:"digits=.4"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"ok", "12345.1234", false},
		{"ok_no_frac", "999999", false},
		{"fail", "1.12345", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Rate: decimal.RequireFromString(tt.val)}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%s) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDigitsZeroInt(t *testing.T) {
	TestRegisterDigitsValidation(t)

	type testStruct struct {
		Ratio decimal.Decimal `v:"digits=0.2"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"ok_zero", "0.12", false},
		{"ok_zero_exact", "0.00", false},
		{"fail_has_int", "1.12", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Ratio: decimal.RequireFromString(tt.val)}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%s) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestDigitsTranslation(t *testing.T) {
	TestRegisterDigitsValidation(t)

	type testStruct struct {
		Price decimal.Decimal `v:"digits=10.2"`
	}

	s := testStruct{Price: decimal.RequireFromString("12345678901.456")}
	err := validate.Struct(s)
	if err == nil {
		t.Fatal("expected error")
	}
	errs := err.(validator.ValidationErrors)
	msg := errs[0].Translate(zhTranslator)
	if !strings.Contains(msg, "整数部分最多10位") || !strings.Contains(msg, "小数部分最多2位") {
		t.Errorf("unexpected translation: %s", msg)
	}
}

func TestValidateDigitsZeroFrac(t *testing.T) {
	TestRegisterDigitsValidation(t)

	type testStruct struct {
		Amount decimal.Decimal `v:"digits=2.0"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"ok_int", "99", false},
		{"ok_int_with_trailing_zero", "99.0", false},
		{"ok_zero", "0", false},
		{"fail_has_frac", "99.1", true},
		{"fail_int_too_long", "123", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Amount: decimal.RequireFromString(tt.val)}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%s) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}
