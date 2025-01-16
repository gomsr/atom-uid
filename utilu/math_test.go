package utilu

import "testing"

func TestNextPowerOfTwo(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "true1", args: args{n: 1}, want: 1},
		{name: "true2", args: args{n: 2}, want: 2},
		{name: "true3", args: args{n: 3}, want: 4},
		{name: "true4", args: args{n: 4}, want: 4},
		{name: "true5", args: args{n: 5}, want: 8},
		{name: "true6", args: args{n: 16}, want: 16},
		{name: "true7", args: args{n: 17}, want: 32},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NextPowerOfTwo(tt.args.n); got != tt.want {
				t.Errorf("NextPowerOfTwo() = %v, want %v", got, tt.want)
			}
		})
	}
}
