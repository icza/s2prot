package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// limit of build spec go file to be considered a non-duplicate
const nonDuplicateLimit = 5000

func TestMap(t *testing.T) {
	// Each *.go file should add 1 entry to one of the the maps
	// If it's "big", then to Builds, else to Duplicates
	folder := "."
	fis, err := ioutil.ReadDir(folder)
	if err != nil {
		t.Errorf("Can't read folder: %s, error: %v", folder, err)
		return
	}
	count := 0
	for _, fi := range fis {
		name := fi.Name()
		if fi.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}

		buildnum, err := strconv.Atoi(name[:len(name)-len(".go")])
		if err != nil {
			continue // Not a protocol build file
		}

		if fi.Size() > nonDuplicateLimit {
			if Builds[buildnum] == "" {
				t.Errorf("Found 'big' build file %s but no matching entry in Builds map!", name)
			}
			if Duplicates[buildnum] != 0 {
				t.Errorf("Found 'big' build file %s and unwanted matching entry in Duplicates map!", name)
			}
		} else {
			if Duplicates[buildnum] == 0 {
				t.Errorf("Found 'small' build file %s but no matching entry in Duplicates map!", name)
			}
			if Builds[buildnum] != "" {
				t.Errorf("Found 'small' build file %s and unwanted matching entry in Builds map!", name)
			}

			if Builds[Duplicates[buildnum]] == "" {
				t.Errorf("There is no matching entry in Builds map for the original base build %d found in 'small' build file %s!", Duplicates[buildnum], name)
			}
		}

		count++
	}

	if ll := len(Builds) + len(Duplicates); ll != count {
		t.Errorf("Found %d protocol build files, but maps have only %d entries!", count, ll)
	}
}

func TestDuplicates(t *testing.T) {
	// Find entries whose values are equal
	// If there are, they are next to each other
	var keys []int

	for k := range Builds {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	for i, key := range keys {
		// Compare to previous ones, find the oldest same:
		for j, key2 := range keys {
			if j == i {
				break
			}
			if Builds[key2] != Builds[key] {
				continue
			}

			// They are the same, the latter should only refer to the oldest same!
			name := fmt.Sprintf("%d.go", key)
			fi, err := os.Stat(name)
			if err != nil {
				t.Errorf("Can't inspect build spec %s: %v", name, err)
				continue
			}
			// If the latter is bigger than a limit, then it contains the duplicate and not just a reference:
			if fi.Size() > 5000 {
				t.Errorf("Builds %d and %d are the same, and so the newer should only contain a reference!\n", key, key2)
			}
			break // no need to search further matches (got the oldest)
		}
	}
}
