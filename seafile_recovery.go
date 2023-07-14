package main

import (
  "log"
  "github.com/docopt/docopt-go"
)

func main() {
  usage := `Seafile Recovery.

Usage:
  seafile_recovery [--storage=<sto>] head <repoid>
  seafile_recovery [--storage=<sto>] ls <repoid> (--dir=<dirid> | --file=<fileid>)
  seafile_recovery [--storage=<sto>] cp <repoid> (--dir=<dirid> | --file=<fileid>) <dest>
  seafile_recovery [--storage=<sto>] s3 <repoid> (--dir=<dirid> | --file=<pathid>) <dest>
  seafile_recovery s3del <dest>
  seafile_recovery (-h | --help)

Options:
  -h --help        Show this screen
  --storage=<sto>  Set Seafile storage path [default: ./storage]
  --dir=<dirid>    Seafile Directory ID, can be obtained from commits as RootID
  --file=<fileid>  Seafile File ID, can be obtained through ls
`

  config := new(configCollect)
  opts, err := docopt.ParseDoc(usage)
  if err != nil { log.Fatal(err) }
  opts.Bind(config)

  if !config.S3Del {
    checkRootFolder(config.Storage)
  }
  rexists := repoExistsIn(config.Storage, config.RepoId)

  if config.Head {
    if !rexists["commits"] { log.Fatal("No commits folder found for repo ",config.RepoId) }
    cmdHead(config)
  } else if config.Ls {
    if !rexists["fs"] { log.Fatal("No fs folder found for repo ",config.RepoId) }

    if len(config.DirId) > 0 {
      cmdLs(config)
    } else  {
      cmdInfo(config)
    }
  } else if config.Cp {
    if !rexists["fs"] { log.Fatal("No fs folder found for repo ",config.RepoId) }
    if !rexists["blocks"] { log.Fatal("No blocks folder found for repo ",config.RepoId) }

    if len(config.DirId) > 0 {
      cmdCpDir(config)
    } else {
      cmdCpFile(config)
    }
  } else if config.S3 {
    if !rexists["fs"] { log.Fatal("No fs folder found for repo ",config.RepoId) }
    if !rexists["blocks"] { log.Fatal("No blocks folder found for repo ",config.RepoId) }

    if len(config.DirId) > 0 {
      cmdS3Dir(config)
    } else {
      cmdS3File(config)
    }
  } else if config.S3Del{
    cmdS3Del(config)
  } else {
    log.Fatal("This command is not implemented")
  }
}
