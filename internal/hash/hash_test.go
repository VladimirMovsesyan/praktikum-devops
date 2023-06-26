package hash

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGet(t *testing.T) {
	key := "superSecretKey"
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "numbers",
			data: "31231212312",
			want: "7414d32b12b5a058fab4d9e9d3be2d9250e42ae9c9fc7e2f6b0a9995f95c3dae",
		},
		{
			name: "alphas",
			data: "fdasfasdfasfdsafieowfajoewafjioewfh",
			want: "731bbf35ebe9b7af01be1d1b2c15343f795281440af493c057a880629b89fd11",
		},
		{
			name: "random",
			data: "gfa3782g38gfwa73fg8a32fegfuz0opq",
			want: "998ac747cc6fff01c55a039726ed42c109684b0a239f441b34244ccd539389b6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, Get(tt.data, key))
		})
	}
}

func TestValid(t *testing.T) {
	key := "superSecretKey"
	tests := []struct {
		name string
		hash string
		data string
		want bool
	}{
		{
			name: "numbers",
			hash: "7414d32b12b5a058fab4d9e9d3be2d9250e42ae9c9fc7e2f6b0a9995f95c3dae",
			data: "31231212312",
			want: true,
		},
		{
			name: "alphas",
			hash: "731bbf35ebe9b7af01be1d1b2c15343f795281440af493c057a880629b89fd11",
			data: "fdasfasdfasfdsafieowfajoewafjioewfh",
			want: true,
		},
		{
			name: "random",
			hash: "998ac747cc6fff01c55a039726ed42c109684b0a239f441b34244ccd539389b6",
			data: "gfa3782g38gfwa73fg8a32fegfuz0opq",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, Valid(tt.hash, tt.data, key))
		})
	}
}
