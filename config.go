package main

type configCollect struct {
  Head bool `docopt:"head"`
  Ls bool `docopt:"ls"`
  Info bool `docopt:"info"`
  Cp bool `docopt:"cp"`
  S3 bool `docopt:"s3"`
  S3Del bool `docopt:"s3del"`
  Storage string `docopt:"--storage"`
  DirId string `docopt:"--dir"`
  FileId string `docopt:"--file"`
  RepoId string `docopt:"<repoid>"`
  Dest string `docopt:"<dest>"`
}
