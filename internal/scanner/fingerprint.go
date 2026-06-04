package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// Fingerprint produces a stable, deterministic hash of the inputs that
// describe a project state. The result is the first 32 hex characters of
// the SHA-256 digest (16 bytes) and is intended to be compared across
// scans, not used as a security identifier.
//
// Inputs, in order:
//
//  1. The project root absolute path.
//  2. The sorted list of relative file paths with their sizes.
//  3. The mtime (RFC3339Nano) of each package manager evidence file.
//  4. The total count of files indexed.
func Fingerprint(root string, files []ScannedFile, packageManagers []PackageManagerHit) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write([]byte("root:"))
	hash.Write([]byte(absRoot))
	hash.Write([]byte{'\n'})

	paths := make([]string, 0, len(files))
	pathToSize := map[string]int64{}
	for _, file := range files {
		paths = append(paths, file.Path)
		pathToSize[file.Path] = file.Size
	}
	sort.Strings(paths)
	hash.Write([]byte("files:"))
	hash.Write([]byte(strconv.Itoa(len(paths))))
	hash.Write([]byte{'\n'})
	for _, p := range paths {
		hash.Write([]byte(p))
		hash.Write([]byte{'='})
		hash.Write([]byte(strconv.FormatInt(pathToSize[p], 10)))
		hash.Write([]byte{'\n'})
	}

	hash.Write([]byte("package_managers:"))
	hash.Write([]byte(strconv.Itoa(len(packageManagers))))
	hash.Write([]byte{'\n'})
	for _, pm := range packageManagers {
		hash.Write([]byte(pm.Name))
		hash.Write([]byte{'='})
		hash.Write([]byte(pm.Evidence))
		hash.Write([]byte{'='})
		mt, err := readMtimeNano(filepath.Join(absRoot, pm.Evidence))
		if err != nil {
			mt = ""
		}
		hash.Write([]byte(mt))
		hash.Write([]byte{'\n'})
	}

	digest := hash.Sum(nil)
	return hex.EncodeToString(digest)[:32], nil
}

func readMtimeNano(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return info.ModTime().UTC().Format(time.RFC3339Nano), nil
}
