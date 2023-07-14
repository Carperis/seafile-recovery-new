package main

import (
  "encoding/json"
  "fmt"
  "log"
  "time"
  "io/ioutil"
  "path/filepath"
)

type Empty struct{}
var empty Empty

type RepoCommits struct {
  Config *configCollect
  CommitDesc map[string]string
  CommitContent map[string]*CommitNode
  Root map[*CommitNode]Empty
  Leafs map[*CommitNode]Empty
  Head *CommitNode
}

type CommitNode struct {
  Id string
  Parents map[*CommitNode]Empty
  Children map[*CommitNode]Empty
  Content Commit
}

type Commit struct {
  CommitId *string `json:"commit_id"`
  RootId *string `json:"root_id"`
  RepoId *string `json:"repo_id"`
  CreatorName *string `json:"creator_name"`
  Creator *string `json:"creator"`
  Description *string `json:"description"`
  Ctime int64 `json:"ctime"`
  ParentId *string `json:"parent_id"`
  SecondParentId *string `json:"second_parent_id"`
  RepoName *string `json:"repo_name"`
  RepoDesc *string `json:"repo_desc"`
  RepoCategory *string `json:"repo_category"`
  NoLocalHistory int `json:"no_local_history"`
  Version int `json:"version"`
}

func nilable(s *string) string {
  if s == nil {
    return "<nil>"
  } else {
    return *s
  }
}

func (c* Commit) String() string {
  return fmt.Sprintf(`RootId: %v
CreatorName: %v
Creator: %v
Description: %v
Ctime: %v
RepoName: %v
RepoDesc: %v
`, nilable(c.RootId), nilable(c.CreatorName), nilable(c.Creator), nilable(c.Description), time.Unix(c.Ctime, 0), nilable(c.RepoName), nilable(c.RepoDesc))
}

func NewRepoCommits (config *configCollect) *RepoCommits {
  rc := new(RepoCommits)
  rc.Config =  config
  rc.CommitDesc = make(map[string]string)
  rc.CommitContent = make(map[string]*CommitNode)
  rc.Root = make(map[*CommitNode]Empty)
  rc.Leafs = make(map[*CommitNode]Empty)
  return rc
}

func (rc* RepoCommits) CollectDescs() {
  baseFolder := filepath.Join(rc.Config.Storage, "commits", rc.Config.RepoId)
  layer1, err := ioutil.ReadDir(baseFolder)
  if err != nil { log.Fatal(err) }
  for _, l1 := range layer1 {
    intFolder := filepath.Join(baseFolder, l1.Name())
    layer2, err := ioutil.ReadDir(intFolder)
    if err != nil { log.Fatal(err) }
    for _, l2 := range layer2 {
      commitId := l1.Name() + l2.Name()
      commitPath := filepath.Join(intFolder, l2.Name())
      rc.CommitDesc[commitId] = commitPath
    }
  }
}

func NewCommitNode(c Commit, id string) *CommitNode {
  cn := new(CommitNode)
  cn.Content = c
  cn.Id = id
  cn.Parents = make(map[*CommitNode]Empty)
  cn.Children = make(map[*CommitNode]Empty)

  return cn
}

func (rc* RepoCommits) CollectContent() {
  for id, path := range rc.CommitDesc {
    data, err := ioutil.ReadFile(path)
    if err != nil { log.Fatal(err) }

    var c Commit;
    err = json.Unmarshal(data, &c)
    if err != nil {
      log.Fatal(err)
    }
    rc.CommitContent[id] = NewCommitNode(c,id)
  }
}

func (rc* RepoCommits) BuildGraph() map[*CommitNode]Empty {
  for _, cn := range rc.CommitContent {
    if cn.Content.ParentId == nil && cn.Content.SecondParentId == nil {
      rc.Root[cn] = empty
    }

    if cn.Content.ParentId != nil {
      parentCn := rc.CommitContent[*cn.Content.ParentId]
      cn.Parents[parentCn] = empty
      parentCn.Children[cn] = empty
    }

    if cn.Content.SecondParentId != nil {
      parentCn := rc.CommitContent[*cn.Content.SecondParentId]
      cn.Parents[parentCn] = empty
      parentCn.Children[cn] = empty
    }
  }

  if len(rc.Root) == 0 {
    log.Fatal("Root commit has not been found")
  }

  return rc.Root
}

func (rc* RepoCommits) FindLeafs() map[*CommitNode]Empty {
  toProcess := make(map[*CommitNode]Empty)

  for cn, _ := range rc.Root {
    toProcess[cn] = empty
  }

  for i := 0; i < len(rc.CommitContent) && len(toProcess) > 0; i++ {
    nextToProcess := make(map[*CommitNode]Empty)
    for cn, _ := range toProcess {
      for ccn, _ := range cn.Children {
        nextToProcess[ccn] = empty
      }
      if len(cn.Children) == 0 {
        rc.Leafs[cn] = empty
      }
    }

    toProcess = nextToProcess
  }

  if len(rc.Leafs) == 0 { log.Fatal("No leafs have been found") }
  return rc.Leafs
}

func (rc* RepoCommits) ChooseHead() *CommitNode {
  for cn, _ := range rc.Leafs {
    if rc.Head == nil { 
      rc.Head = cn 
    } else if rc.Head.Content.Ctime < cn.Content.Ctime { 
      rc.Head = cn 
    }
  }
  if rc.Head == nil { log.Fatal("No HEAD has been found") }
  return rc.Head
}
