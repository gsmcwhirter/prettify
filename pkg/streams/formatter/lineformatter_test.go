package formatter

import (
	"reflect"
	"testing"
)

func Test_formatResultStrings(t *testing.T) {
	t.Parallel()
	type args struct {
		resStrings    []string
		separatorType gJSONOutputType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "csv",
			args: args{
				resStrings:    []string{"a", "b", "c"},
				separatorType: csv,
			},
			want: "a,b,c",
		},
		{
			name: "tsv",
			args: args{
				resStrings:    []string{"a", "b", "c"},
				separatorType: tsv,
			},
			want: "a\tb\tc",
		},
		{
			name: "space",
			args: args{
				resStrings:    []string{"a", "b", "c"},
				separatorType: space,
			},
			want: "a b c",
		},
		{
			name: "nsv",
			args: args{
				resStrings:    []string{"a", "b", "c"},
				separatorType: nlsv,
			},
			want: "a\nb\nc",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatResultStrings(tt.args.resStrings, tt.args.separatorType); got != tt.want {
				t.Errorf("formatResultStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatResultBytes(t *testing.T) {
	t.Parallel()
	type args struct {
		resBytes      [][]byte
		separatorType gJSONOutputType
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "csv",
			args: args{
				resBytes:      [][]byte{[]byte("a"), []byte("b"), []byte("c")},
				separatorType: csv,
			},
			want: []byte("a,b,c"),
		},
		{
			name: "tsv",
			args: args{
				resBytes:      [][]byte{[]byte("a"), []byte("b"), []byte("c")},
				separatorType: tsv,
			},
			want: []byte("a\tb\tc"),
		},
		{
			name: "space",
			args: args{
				resBytes:      [][]byte{[]byte("a"), []byte("b"), []byte("c")},
				separatorType: space,
			},
			want: []byte("a b c"),
		},
		{
			name: "nsv",
			args: args{
				resBytes:      [][]byte{[]byte("a"), []byte("b"), []byte("c")},
				separatorType: nlsv,
			},
			want: []byte("a\nb\nc"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatResultBytes(tt.args.resBytes, tt.args.separatorType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formatResultBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getGJSONPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		gJSONPath string
		wantPaths []string
		wantType  gJSONOutputType
	}{
		{
			name:      "default type",
			gJSONPath: "foo, bar",
			wantPaths: []string{"foo", "bar"},
			wantType:  nlsv,
		},
		{
			name:      "csv type",
			gJSONPath: "foo, bar,|@csv",
			wantPaths: []string{"foo", "bar"},
			wantType:  csv,
		},
		{
			name:      "tsv type",
			gJSONPath: "foo, bar,|@tsv",
			wantPaths: []string{"foo", "bar"},
			wantType:  tsv,
		},
		{
			name:      "spacesep type",
			gJSONPath: "foo, bar,|@ssv",
			wantPaths: []string{"foo", "bar"},
			wantType:  space,
		},
		{
			name:      "nlsv type",
			gJSONPath: "foo, bar, |@nlsv",
			wantPaths: []string{"foo", "bar"},
			wantType:  nlsv,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotPaths, gotType := getGJSONPaths(tt.gJSONPath)
			if !reflect.DeepEqual(gotPaths, tt.wantPaths) {
				t.Errorf("LinePrinter.getGJSONPaths() got paths = %v, want %v", gotPaths, tt.wantPaths)
			}
			if !reflect.DeepEqual(gotType, tt.wantType) {
				t.Errorf("LinePrinter.getGJSONPaths() got type = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestFormatLine(t *testing.T) {
	t.Parallel()
	type args struct {
		line            string
		formatSelectors string
		prettyFmt       bool
		withColor       bool
		sortKeys        bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default type",
			args: args{
				line:            `{"foo": 1, "bar": "foo"}`,
				formatSelectors: "foo, bar",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: "1\nfoo",
		},
		{
			name: "csv type",
			args: args{
				line:            `{"foo": 1, "bar": "foo"}`,
				formatSelectors: "foo, bar, |@csv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: "1,foo",
		},
		{
			name: "tsv type",
			args: args{
				line:            `{"foo": 1, "bar": "foo"}`,
				formatSelectors: "foo, bar, |@tsv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: "1\tfoo",
		},
		{
			name: "spacesep type",
			args: args{
				line:            `{"foo": 1, "bar": "foo"}`,
				formatSelectors: "foo, bar, |@ssv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: "1 foo",
		},
		{
			name: "nlsv type",
			args: args{
				line:            `{"foo": 1, "bar": "foo"}`,
				formatSelectors: "foo, bar, |@nlsv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: "1\nfoo",
		},
		{
			name: "missing fields",
			args: args{
				line:            `{"foo": 1, "bar": "foo"}`,
				formatSelectors: "foo, baz, bar, |@tsv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: "1\t\tfoo",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := FormatLine(tt.args.line, tt.args.formatSelectors, tt.args.prettyFmt, tt.args.withColor, tt.args.sortKeys); got != tt.want {
				t.Errorf("FormatLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatLineBytes(t *testing.T) {
	t.Parallel()
	type args struct {
		line            []byte
		formatSelectors string
		prettyFmt       bool
		withColor       bool
		sortKeys        bool
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "default type",
			args: args{
				line:            []byte(`{"foo": 1, "bar": "foo"}`),
				formatSelectors: "foo, bar",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: []byte("1\nfoo"),
		},
		{
			name: "csv type",
			args: args{
				line:            []byte(`{"foo": 1, "bar": "foo"}`),
				formatSelectors: "foo, bar, |@csv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: []byte("1,foo"),
		},
		{
			name: "tsv type",
			args: args{
				line:            []byte(`{"foo": 1, "bar": "foo"}`),
				formatSelectors: "foo, bar, |@tsv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: []byte("1\tfoo"),
		},
		{
			name: "spacesep type",
			args: args{
				line:            []byte(`{"foo": 1, "bar": "foo"}`),
				formatSelectors: "foo, bar, |@ssv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: []byte("1 foo"),
		},
		{
			name: "nlsv type",
			args: args{
				line:            []byte(`{"foo": 1, "bar": "foo"}`),
				formatSelectors: "foo, bar, |@nlsv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: []byte("1\nfoo"),
		},
		{
			name: "missing fields",
			args: args{
				line:            []byte(`{"foo": 1, "bar": "foo"}`),
				formatSelectors: "foo, baz, bar, |@tsv",
				prettyFmt:       false,
				withColor:       false,
				sortKeys:        false,
			},
			want: []byte("1\t\tfoo"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := FormatLineBytes(tt.args.line, tt.args.formatSelectors, tt.args.prettyFmt, tt.args.withColor, tt.args.sortKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatLineBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrettyLine(t *testing.T) {
	t.Parallel()
	type args struct {
		line      string
		withColor bool
		sortKeys  bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no color",
			args: args{
				line:      `{"b": 1, "a": "foo"}`,
				withColor: false,
				sortKeys:  false,
			},
			want: "{\n  \"b\": 1,\n  \"a\": \"foo\"\n}\n",
		},
		{
			name: "color",
			args: args{
				line:      `{"b": 1, "a": "foo"}`,
				withColor: true,
				sortKeys:  false,
			},
			// blue: \x1b\x5b\x39\x34\x6d
			// yellow: \x1b\x5b\x39\x33\x6d
			// green: \x1b\x5b\x39\x32\x6d
			// reset: \x1b\x5b\x30\x6d
			want: "{\n  \x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d,\n  \x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d\n}\n",
		},
		{
			name: "no color sort",
			args: args{
				line:      `{"b": 1, "a": "foo"}`,
				withColor: false,
				sortKeys:  true,
			},
			want: "{\n  \"a\": \"foo\",\n  \"b\": 1\n}\n",
		},
		{
			name: "color sort",
			args: args{
				line:      `{"b": 1, "a": "foo"}`,
				withColor: true,
				sortKeys:  true,
			},
			// blue: \x1b\x5b\x39\x34\x6d
			// yellow: \x1b\x5b\x39\x33\x6d
			// green: \x1b\x5b\x39\x32\x6d
			// reset: \x1b\x5b\x30\x6d
			want: "{\n  \x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d,\n  \x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d\n}\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := PrettyLine(tt.args.line, tt.args.withColor, tt.args.sortKeys); got != tt.want {
				t.Errorf("PrettyLine() =\n%v\n(%v), want\n%v\n(%v)", got, []byte(got), tt.want, []byte(tt.want))
			}
		})
	}
}

func TestPrettyLineBytes(t *testing.T) {
	t.Parallel()
	type args struct {
		line      []byte
		withColor bool
		sortKeys  bool
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "no color",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: false,
				sortKeys:  false,
			},
			want: []byte("{\n  \"b\": 1,\n  \"a\": \"foo\"\n}\n"),
		},
		{
			name: "color",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: true,
				sortKeys:  false,
			},
			// blue: \x1b\x5b\x39\x34\x6d
			// yellow: \x1b\x5b\x39\x33\x6d
			// green: \x1b\x5b\x39\x32\x6d
			// reset: \x1b\x5b\x30\x6d
			want: []byte("{\n  \x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d,\n  \x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d\n}\n"),
		},
		{
			name: "no color sort",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: false,
				sortKeys:  true,
			},
			want: []byte("{\n  \"a\": \"foo\",\n  \"b\": 1\n}\n"),
		},
		{
			name: "color sort",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: true,
				sortKeys:  true,
			},
			// blue: \x1b\x5b\x39\x34\x6d
			// yellow: \x1b\x5b\x39\x33\x6d
			// green: \x1b\x5b\x39\x32\x6d
			// reset: \x1b\x5b\x30\x6d
			want: []byte("{\n  \x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d,\n  \x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d\n}\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := PrettyLineBytes(tt.args.line, tt.args.withColor, tt.args.sortKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrettyLineBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUglyLineBytes(t *testing.T) {
	t.Parallel()
	type args struct {
		line      []byte
		withColor bool
		sortKeys  bool
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "no color",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: false,
				sortKeys:  false,
			},
			want: []byte("{\"b\":1,\"a\":\"foo\"}"),
		},
		{
			name: "color",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: true,
				sortKeys:  false,
			},
			// blue: \x1b\x5b\x39\x34\x6d
			// yellow: \x1b\x5b\x39\x33\x6d
			// green: \x1b\x5b\x39\x32\x6d
			// reset: \x1b\x5b\x30\x6d
			want: []byte("{\x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d:\x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d,\x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d:\x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d}"),
		},
		{
			name: "no color sort",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: false,
				sortKeys:  true,
			},
			want: []byte("{\"a\":\"foo\",\"b\":1}"),
		},
		{
			name: "color sort",
			args: args{
				line:      []byte(`{"b": 1, "a": "foo"}`),
				withColor: true,
				sortKeys:  true,
			},
			// blue: \x1b\x5b\x39\x34\x6d
			// yellow: \x1b\x5b\x39\x33\x6d
			// green: \x1b\x5b\x39\x32\x6d
			// reset: \x1b\x5b\x30\x6d
			want: []byte("{\x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d:\x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d,\x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d:\x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d}"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := UglyLineBytes(tt.args.line, tt.args.withColor, tt.args.sortKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UglyLineBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
