package generator

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	code, err := Generate(DefaultLength)

	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(code) != DefaultLength {
		t.Fatalf(
			"expected length %d, got %d",
			DefaultLength,
			len(code),
		)
	}
}

func TestGenerateInvalidLength(t *testing.T) {
	_, err := Generate(0)

	if err == nil {
		t.Fatal("expected error for zero length")
	}
}

func TestGenerateCustomLength(t *testing.T) {
	const length = 10

	code, err := Generate(length)

	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(code) != length {
		t.Fatalf("expected length %d, got %d", length, len(code))
	}
}

func TestGenerateCharacters(t *testing.T) {
	code, err := Generate(DefaultLength)

	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	for _, ch := range code {
		if !strings.ContainsRune(alphabet, ch) {
			t.Fatalf("generated invalid character: %q", ch)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
    generated := make(map[string]bool)

    for i := 0; i < 1000; i++ {
        code, err := Generate(DefaultLength)

        if err != nil {
            t.Fatalf("Generate() returned error: %v", err)
        }

        if generated[code] {
            t.Fatalf("duplicate short code generated: %s", code)
        }

        generated[code] = true
    }
}