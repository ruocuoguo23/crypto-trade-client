package rpc

import "testing"

func TestCamelCaseName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Convert TwoWords to twoWords",
			args: args{name: "TwoWords"},
			want: "twoWords",
		},
		{
			name: "twoWords will not change",
			args: args{name: "twoWords"},
			want: "twoWords",
		},
		{
			name: "Convert One to one",
			args: args{name: "One"},
			want: "one",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CamelCaseName(tt.args.name); got != tt.want {
				t.Errorf("CamelCaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnakeCaseName(t *testing.T) {
	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Convert TwoWords to two_words",
			args: args{name: "TwoWords"},
			want: "two_words",
		},
		{
			name: "Convert twoWords two_words",
			args: args{name: "twoWords"},
			want: "two_words",
		},
		{
			name: "Convert One to one",
			args: args{name: "One"},
			want: "one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnakeCaseName(tt.args.name); got != tt.want {
				t.Errorf("SnakeCaseName() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
