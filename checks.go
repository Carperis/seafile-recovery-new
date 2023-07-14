package main

import(
  "log"
  "os"
  "path/filepath"
)

func checkRootFolder(storage string) {
  checked_folders := map[string]bool{"fs": false, "commits": false, "blocks": false}
  for f, _ := range checked_folders {
    if info, err := os.Stat(filepath.Join(storage, f));  err == nil && info.IsDir() {
      checked_folders[f] = true
    }
  }

  for path, seen := range checked_folders {
    if !seen { log.Fatal("Folder ", path, " is required but not present!") }
  }
}

func repoExistsIn(storage string, repoId string) map[string]bool {
  exists_in := map[string]bool{"fs": false, "commits": false, "blocks": false}

  for storageType, _ := range exists_in {
    if info, err := os.Stat(filepath.Join(storage, storageType, repoId)); err == nil && info.IsDir() {
      exists_in[storageType] = true
    }
  }
  return exists_in
}
