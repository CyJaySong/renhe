package custom

import (
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
)

var onceTestRegisterEmojiValidation sync.Once

func TestRegisterEmojiValidation(t *testing.T) {
	initValidator()
	onceTestRegisterEmojiValidation.Do(func() {
		if err := RegisterEmojiValidation(validate, zhTranslator); err != nil {
			t.Errorf("unexpected register: %s", err)
		}
	})
}

func TestValidateNoEmoji(t *testing.T) {
	TestRegisterEmojiValidation(t)

	type testStruct struct {
		Text string `v:"noemoji"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"empty", "", false},
		{"pure_ascii", "hello world", false},
		{"chinese", "你好世界", false},
		{"mixed", "hello你好123", false},
		{"punctuation", "hello, world!", false},
		{"has_smile", "hello😀", true},
		{"has_heart", "I❤️you", true},
		{"only_emoji", "🎉", true},
		{"emoji_in_middle", "abc🔥def", true},
		{"thumbs_up", "good👍", true},
		{"flag_emoji", "🇨🇳", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Text: tt.val}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("noemoji(%q) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestNoEmojiTranslation(t *testing.T) {
	TestRegisterEmojiValidation(t)

	type testStruct struct {
		Text string `v:"noemoji"`
	}

	s := testStruct{Text: "hello😀"}
	err := validate.Struct(s)
	if err == nil {
		t.Fatal("expected error")
	}
	errs := err.(validator.ValidationErrors)
	msg := errs[0].Translate(zhTranslator)
	if !strings.Contains(msg, "表情符号") {
		t.Errorf("unexpected translation: %s", msg)
	}
}
