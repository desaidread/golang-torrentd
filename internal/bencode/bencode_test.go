package bencode

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeInt(t *testing.T) {
	// шаг 1: описываем кейсы — вход и ожидаемый результат
	tests := []struct {
		name    string // название кейса, для читаемости в выводе
		input   string
		want    int64
		wantErr bool // ожидаем ли мы ошибку
	}{
		{name: "positive", input: "i42e", want: 42, wantErr: false},
		{name: "negative", input: "i-3e", want: -3, wantErr: false},
		{name: "zero", input: "i0e", want: 0, wantErr: false},
		{name: "invalid, no digits", input: "ie", want: 0, wantErr: true},
	}

	// шаг 2: прогоняем каждый кейс
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(bufio.NewReader(strings.NewReader(tt.input)))
			got, err := d.decodeInt()

			// проверяем, совпало ли наличие ошибки с ожиданием
			if (err != nil) != tt.wantErr {
				t.Fatalf("decodeInt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return // если ждали ошибку и она есть — на этом кейс закончен
			}

			// проверяем значение
			if got != tt.want {
				t.Errorf("decodeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []any
	}{
		{name: "two strings", input: "l4:spam4:eggse", want: []any{"spam", "eggs"}},
		{name: "empty list", input: "le", want: []any(nil)},
		{name: "mixed types", input: "l4:spami42ee", want: []any{"spam", int64(42)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(bufio.NewReader(strings.NewReader(tt.input)))
			got, err := d.decodeList()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeList() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecodeDict(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{name: "nested dict", input: "d4:infod6:lengthi100eee", want: map[string]any{
			"info": map[string]any{"length": int64(100)},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(bufio.NewReader(strings.NewReader(tt.input)))
			got, err := d.decodeDict()
			if err != nil {
				t.Fatalf("unxepected error: %v", err)

			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeDict() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecodeString(t *testing.T) {
	test := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "test1", input: "4:spam", want: "spam", wantErr: false},
		{name: "zero", input: "0:", want: "", wantErr: false},
		{name: "unknown", input: "x", want: "", wantErr: true},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(bufio.NewReader(strings.NewReader(tt.input)))
			got, err := d.decodeString()

			if (err != nil) != tt.wantErr {
				t.Fatalf("decodeString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("decodeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	test := []struct {
		name  string
		input string
		want  any
	}{
		{name: "three levels", input: "d3:cow3:mooe", want: map[string]any{"cow": "moo"}},
		{name: "string", input: "4:spam", want: "spam"},
		{name: "int", input: "i42e", want: int64(42)},
		{name: "list", input: "l4:spam4:eggse", want: []any{"spam", "eggs"}},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(bufio.NewReader(strings.NewReader(tt.input)))
			got, err := d.Decode()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
