package main

import ("log")

type LsWalker struct {
  TotalSize uint64
 }
func (lw* LsWalker) onDir(dn *DirNode) {
  log.Println(dn.String())
}
func (lw* LsWalker) onFile(fn *FileNode) {
  lw.TotalSize += uint64(fn.Ent.Size)
  log.Println(fn.String())
}
