package magic_test

import (
	"testing"

	"github.com/omalloc/contrib/v2/magic"
)

func TestFireMagic(t *testing.T) {
	fireMagic := magic.NewFireMagic()
	fireMagic.Trigger()
}
