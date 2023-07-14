package main

import (
  "log"
  "os"
  "path/filepath"
  "io"
)

type CopyWalker struct { }
func (cw* CopyWalker) onDir(dn *DirNode) {
  err := os.MkdirAll(filepath.Join(dn.Config.Dest, dn.AbsolutePath), os.ModePerm)
  if err != nil { log.Fatal(err) }
  log.Println(dn.String())
}
func (cw* CopyWalker) onFile(fn *FileNode) {
  fn.Parse()
  path := filepath.Join(fn.Config.Dest, fn.AbsolutePath)
  file, err := os.Create(path)
  if err != nil { log.Fatal(err) }
  defer file.Close()
  written, err := io.Copy(file, fn)
  if err != nil { log.Fatal(err) }
  if uint64(written) != fn.Elem.FileSize {
    log.Fatal(written, "bytes written,", fn.Elem.FileSize, "bytes expected")
  }
  log.Println(fn.String())
}
