package custom

import (
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
)

var onceTestRegisterHanziValidation sync.Once

func TestRegisterHanziValidation(t *testing.T) {
	initValidator()
	onceTestRegisterHanziValidation.Do(func() {
		if err := RegisterHanziValidation(validate, zhTranslator); err != nil {
			t.Errorf("unexpected register: %s", err)
		}
	})
}

func TestValidateHanzi(t *testing.T) {
	TestRegisterHanziValidation(t)

	type testStruct struct {
		Name string `v:"hanzi"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"pure_hanzi", "你好世界", false},
		{"single_char", "中", false},
		{"empty", "", true},
		{"has_digit", "你好123", true},
		{"has_letter", "你好abc", true},
		{"has_space", "你好 世界", true},
		{"has_dot", "你好·世界", true},
		{"all_ascii", "hello", true},
		{"has_punctuation", "你好！", true},
		{"japanese_kanji", "漢字", false}, // 日文汉字也属于 unicode.Han
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Name: tt.val}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("hanzi(%q) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestHanziTranslation(t *testing.T) {
	TestRegisterHanziValidation(t)

	type testStruct struct {
		Name string `v:"hanzi"`
	}

	s := testStruct{Name: "abc"}
	err := validate.Struct(s)
	if err == nil {
		t.Fatal("expected error")
	}
	errs := err.(validator.ValidationErrors)
	msg := errs[0].Translate(zhTranslator)
	if !strings.Contains(msg, "汉字") {
		t.Errorf("unexpected translation: %s", msg)
	}
}
