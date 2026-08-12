package encode

import "testing"

const testMixMultiplier = 1234567891

func TestObfuscate(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want int
	}{
		{
			name: "zero id",
			id:   0,
			want: 0,
		},
		{
			name: "id one",
			id:   1,
			want: testMixMultiplier,
		},
		{
			name: "normal id",
			id:   2,
			want: 2469135782,
		},
		{
			name: "32 bit boundary",
			id:   1 << 32,
			want: 0,
		},
		{
			name: "32 bit mask after multiplication",
			id:   (1 << 32) + 5,
			want: 1877872159,
		},
		{
			name: "high bit masked",
			id:   1 << 31,
			want: 2147483648,
		},
		{
			name: "maximum 32 bit value",
			id:   (1 << 32) - 1,
			want: 3060399405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := obfuscate(tt.id, testMixMultiplier)

			if got != tt.want {
				t.Errorf(
					"obfuscate(%d, %d) = %d, want %d",
					tt.id,
					testMixMultiplier,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want string
	}{
		{
			name: "zero",
			id:   0,
			want: "0",
		},
		{
			name: "one",
			id:   1,
			want: "1",
		},
		{
			name: "last numeric character",
			id:   9,
			want: "9",
		},
		{
			name: "first uppercase character",
			id:   10,
			want: "A",
		},
		{
			name: "last uppercase character",
			id:   35,
			want: "Z",
		},
		{
			name: "first lowercase character",
			id:   36,
			want: "a",
		},
		{
			name: "last lowercase character",
			id:   61,
			want: "z",
		},
		{
			name: "base62 rollover",
			id:   62,
			want: "10",
		},
		{
			name: "base62 rollover two",
			id:   63,
			want: "11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Multiplier 1 isolates the Base62 conversion.
			got := EncodeBase62(tt.id, 1)

			if got != tt.want {
				t.Errorf(
					"EncodeBase62(%d, 1) = %q, want %q",
					tt.id,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestEncodeBase62_OnlyContainsValidCharacters(t *testing.T) {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	for id := 0; id < 10000; id++ {
		got := EncodeBase62(id, testMixMultiplier)

		for _, char := range got {
			valid := false

			for _, allowed := range chars {
				if char == allowed {
					valid = true
					break
				}
			}

			if !valid {
				t.Fatalf(
					"EncodeBase62(%d, %d) = %q contains invalid character %q",
					id,
					testMixMultiplier,
					got,
					char,
				)
			}
		}
	}
}

func TestEncodeBase62_Deterministic(t *testing.T) {
	const id = 123456

	first := EncodeBase62(id, testMixMultiplier)
	second := EncodeBase62(id, testMixMultiplier)

	if first != second {
		t.Errorf(
			"EncodeBase62 is not deterministic: first=%q, second=%q",
			first,
			second,
		)
	}
}
