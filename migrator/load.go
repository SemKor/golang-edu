package migrator

import (
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version  int
	Name     string
	UpFile   string
	DownFile string
}

func LoadMigrations(dir string) []Migration {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	mMap := make(map[int]*Migration)

	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		parts := strings.SplitN(name, "_", 2)
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if _, ok := mMap[v]; !ok {
			mMap[v] = &Migration{Version: v}
		}

		if strings.HasSuffix(name, ".up.sql") {
			mMap[v].UpFile = dir + "/" + name
		} else if strings.HasSuffix(name, ".down.sql") {
			mMap[v].DownFile = dir + "/" + name
		}
	}

	migs := make([]Migration, 0, len(mMap))
	for _, m := range mMap {
		migs = append(migs, *m)
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].Version < migs[j].Version })
	return migs
}