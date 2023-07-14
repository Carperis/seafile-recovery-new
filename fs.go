package main

import(
  "compress/zlib"
  "fmt"
  "log"
  "os"
  "io"
  "io/ioutil"
  "encoding/json"
  "syscall"
  "path/filepath"
)

// Imported from https://github.com/haiwen/seafile-server/blob/master/fileserver/fsmgr/fsmgr.go
// License AGPLv3

const emptyId = "0000000000000000000000000000000000000000"

//SeafDir is a dir object
type DirElem struct {
  Version int           `json:"version"`
  DirType int           `json:"type,omitempty"`
  DirId   string        `json:"dir_id,omitempty"`
  Entries []*DirEnt     `json:"dirents"`
}

// SeafDirent is a dir entry object
type DirEnt struct {
  Mode     uint32 `json:"mode"`
  Id       string `json:"id"`
  Name     string `json:"name"`
  Mtime    int64  `json:"mtime"`
  Modifier string `json:"modifier"`
  Size     int64  `json:"size"`
}

type FileElem struct {
  Version  int      `json:"version"`
  FileType int      `json:"type,omitempty"`
  FileId   string   `json:"file_id,omitempty"`
  FileSize uint64   `json:"size"`
  BlkIds   []string `json:"block_ids"`
}

type DirNode struct {
  Config *configCollect
  Ent *DirEnt
  Elem DirElem
  AbsolutePath string
}

type FileNode struct {
  Config *configCollect
  Ent *DirEnt
  Elem FileElem
  AbsolutePath string

  ReadBytes uint64
  CurrentBlock *os.File
  RemainingBlocks []string
}

func (fe* FileElem) String() string {
  return fmt.Sprintf("Size: %v, Blocks: %v", fe.FileSize, len(fe.BlkIds))
}

type TreeObserver interface {
  onDir(dir *DirNode)
  onFile(file *FileNode)
}

func expandId(config *configCollect, userId string) string {
  if len(userId) == len(emptyId) {
    return userId
  }

  if len(userId) < 2 {
    log.Fatal("User ID",userId,"is too short and thus cannot be expanded")
  }

  path := filepath.Join(config.Storage, "fs", config.RepoId, userId[:2])
  files, err := ioutil.ReadDir(path)
  if err != nil { log.Fatal("Unable to read dir", path, "in order to expand ID", userId) }
  for _, f := range files {
    if f.Name()[:len(userId)-2] == userId[2:] {
      return userId[:2] + f.Name()
    }
  }

  log.Fatal("Unable to find", userId[2:],"in",path)
  return emptyId
}

func NewEntryNode(config *configCollect) *DirNode {
  entryNode := new(DirNode)
  entryNode.Ent = new(DirEnt)
  entryNode.Ent.Id = expandId(config, config.DirId)
  entryNode.Ent.Name = ""
  entryNode.AbsolutePath = "/"
  entryNode.Config = config

  return entryNode
}

func NewDirNode(parent *DirNode, ent *DirEnt) *DirNode {
  dn := new(DirNode)
  dn.Ent = ent
  dn.Config = parent.Config
  dn.AbsolutePath = parent.AbsolutePath + ent.Name + "/"

  return dn
}

func (dn* DirNode) String() string {
  return fmt.Sprintf("%v %v", dn.Ent.Id[:6], dn.AbsolutePath)
}

func (fn* FileNode) String() string {
  return fmt.Sprintf("%v %v", fn.Ent.Id[:6], fn.AbsolutePath)
}

func NewFileNode(parent *DirNode, ent *DirEnt) *FileNode {
  fn := new (FileNode)
  fn.Ent = ent
  fn.Config = parent.Config
  fn.AbsolutePath = parent.AbsolutePath + ent.Name

  return fn
}

func NewEntryFileNode(config *configCollect) *FileNode {
  fn := new (FileNode)
  fn.Ent = new(DirEnt)
  fn.Ent.Id = expandId(config, config.FileId)
  fn.Config = config

  return fn
}


func (dn* DirNode) Parse() {
  path := filepath.Join(dn.Config.Storage, "fs", dn.Config.RepoId, dn.Ent.Id[:2], dn.Ent.Id[2:])

  file, err := os.Open(path)
  if err != nil { log.Fatal(err) }
  defer file.Close()

  zfile, err := zlib.NewReader(file)
  if err != nil { log.Fatal(err) }

  jdec := json.NewDecoder(zfile)
  err = jdec.Decode(&dn.Elem)
  if err != nil { log.Fatal(err) }

}

func (fn* FileNode) Parse() {
  path := filepath.Join(fn.Config.Storage, "fs", fn.Config.RepoId, fn.Ent.Id[:2], fn.Ent.Id[2:])

  file, err := os.Open(path)
  if err != nil { log.Fatal(err) }
  defer file.Close()

  zfile, err := zlib.NewReader(file)
  if err != nil { log.Fatal(err) }

  jdec := json.NewDecoder(zfile)
  err = jdec.Decode(&fn.Elem)
  if err != nil { log.Fatal(err) }

  fn.RemainingBlocks = fn.Elem.BlkIds
}

func (fn* FileNode) Read(p []byte) (n int, err error) {
  if fn.CurrentBlock == nil && len(fn.RemainingBlocks) == 0 {
    return 0, io.EOF
  }

  if fn.CurrentBlock == nil {
    blockId := fn.RemainingBlocks[0]
    fn.RemainingBlocks = fn.RemainingBlocks[1:]
    path := filepath.Join(fn.Config.Storage, "blocks", fn.Config.RepoId, blockId[:2], blockId[2:])
    fn.CurrentBlock, err = os.Open(path)
    if err != nil {
      return 0, err
    }
  }

  n, err = fn.CurrentBlock.Read(p)
  fn.ReadBytes += uint64(n)

  if err == io.EOF {
    err = nil
    fn.CurrentBlock.Close()
    fn.CurrentBlock = nil
  }

  return n, err
}

func (dn* DirNode) Children() ([]*DirNode, []*FileNode) {
  folders := make([]*DirNode,0)
  files := make([]*FileNode, 0)

  for _, el := range dn.Elem.Entries {
    if el.Id == emptyId {
      log.Println("[Lost] "+dn.AbsolutePath+el.Name)
    } else if IsDir(el.Mode) {
      folders = append(folders, NewDirNode(dn, el))
    } else if IsRegular(el.Mode) {
      files = append(files, NewFileNode(dn, el))
    } else {
      log.Fatal("Unknown mode", el.Mode, "for", el.Name, "in object id", dn.Ent.Id)
    }
  }

  return folders, files
}

func (dn* DirNode) Walk(tobs TreeObserver) {
  qNode := []*DirNode{dn}
  for len(qNode) > 0 {

    cursor := qNode[0]
    qNode = qNode[1:]
    cursor.Parse()
    dir, file := cursor.Children()
    qNode = append(dir, qNode...)

    tobs.onDir(cursor)
    for _, f := range file {
      tobs.onFile(f)
    }
  }
}

// IsDir check if the mode is dir.
func IsDir(m uint32) bool {
  return (m & syscall.S_IFMT) == syscall.S_IFDIR
}

// IsRegular Check if the mode is regular.
func IsRegular(m uint32) bool {
  return (m & syscall.S_IFMT) == syscall.S_IFREG
}


