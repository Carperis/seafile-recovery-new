package main

import(
  "log"
  "code.cloudfoundry.org/bytefmt"
)

func cmdHead(config *configCollect) {
  rc := NewRepoCommits(config)

  rc.CollectDescs()
  log.Println("Repo contains", len(rc.CommitDesc), "commits")

  rc.CollectContent()
  rc.BuildGraph()
  log.Println("Repo has", len(rc.Root), "sources")
  if len(rc.Root) > 1 {
    for cn, _ := range rc.Root {
      log.Println("Commit "+cn.Id+"\n"+cn.Content.String()) 
    }
  }

  rc.FindLeafs()
  log.Println("Repo has", len(rc.Leafs), "sinks")

  rc.ChooseHead()
  log.Println("Proposing following HEAD:\n"+rc.Head.Content.String())
}


func cmdLs(config *configCollect) {
  en := NewEntryNode(config)
  lw := new(LsWalker)
  en.Walk(lw)
  log.Println("Total size:", bytefmt.ByteSize(lw.TotalSize))
}

func cmdInfo(config *configCollect) {
  en := NewEntryFileNode(config)
  en.Parse()
  log.Println(en.Elem.String())
}

func cmdCpFile(config *configCollect) {
  en := NewEntryFileNode(config)
  cw := new(CopyWalker)
  cw.onFile(en)
}

func cmdCpDir(config *configCollect) {
  en := NewEntryNode(config)
  en.Walk(new(CopyWalker))
}

func cmdS3Dir(config *configCollect) {
  en := NewEntryNode(config)
  en.Walk(NewS3Walker(config))
}

func cmdS3File(config *configCollect) {
  en := NewEntryFileNode(config)
  sw := NewS3Walker(config)
  sw.onFile(en)
}

func cmdS3Del(config *configCollect) {
  sw := NewS3Walker(config)
  sw.onDelete()
}
