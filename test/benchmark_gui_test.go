package test

import (
	"bytes"
	"testing"

	"github.com/dpologdvinity/promptpet/internal/cli"
)

func BenchmarkPromptWithEventLogging(b *testing.B) {
	b.Setenv("PROMPTPET_HOME", b.TempDir())
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cli.Run([]string{"prompt", "--exit=0"}, &buf, &buf)
	}
}
