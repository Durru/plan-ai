package installer

import (
	"os"

	"github.com/plan-ai/plan-ai/internal/atomicfile"
)

// writeFileAtomically is a thin wrapper around atomicfile.WriteFile,
// preserved for the installer's existing call sites.
func writeFileAtomically(path string, data []byte, perm os.FileMode) error {
	return atomicfile.WriteFile(path, data, perm)
}
