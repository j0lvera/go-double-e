package server

import (
	"github.com/j0lvera/go-double-e/internal/testutils"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testutils.SetupTestLogger()

	os.Exit(m.Run())
}
