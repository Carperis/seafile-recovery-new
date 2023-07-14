package main

import (
  "context"
  "log"
  "mime"
  "net/url"
  "path/filepath"
  "strings"

  "github.com/minio/minio-go/v7"
  "github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Walker struct {
  ctx context.Context
  config *configCollect
  mc *minio.Client
  bucket string
  pathPrefix string
}
func NewS3Walker(config *configCollect) *S3Walker {
  sw := new(S3Walker)

  sw.ctx = context.Background()
  sw.config = config

  u, err := url.Parse(config.Dest)
  if err != nil {
    log.Fatal(err)
  }

  if u.Scheme != "s3" { log.Fatal("URL must be of the following form: s3://ACCESS_KEY:SECRET_KEY@ENDPOINT/REGION/BUCKET[/PREFIX") }

  accessKeyID := u.User.Username()
  secretAccessKey, _ := u.User.Password()
  endpoint := u.Host
  splittedPath := strings.SplitN(u.Path, "/", 4)
  if len(splittedPath) < 3 {
    log.Fatal("Bucket or region not found")
  }
  region := splittedPath[1]
  sw.bucket = splittedPath[2]
  if len(splittedPath) > 3 {
    sw.pathPrefix = splittedPath[3]
  }

  useSSL := true

  // Initialize minio client object.
  sw.mc, err = minio.New(endpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
    Secure: useSSL,
    Region: region,
  })
  if err != nil {
    log.Fatal(err)
  }

  //sw.mc.TraceOn(nil)

  return sw
}
func (sw* S3Walker) onDir(dn *DirNode) {
}
func (sw* S3Walker) onFile(fn *FileNode) {
  fn.Parse()
  p := filepath.Join(sw.pathPrefix, fn.AbsolutePath)[1:]
  contentType := mime.TypeByExtension(filepath.Ext(p))
  info, err := sw.mc.PutObject(sw.ctx, sw.bucket, p, fn, int64(fn.Elem.FileSize), minio.PutObjectOptions{ ContentType: contentType })
  if err != nil {
    log.Fatalln(err)
  } else if info.Size != int64(fn.Elem.FileSize) {
    log.Fatal("Uploaded",info.Size,"bytes but expected", fn.Elem.FileSize,"bytes")
  }
  log.Println(fn.String())
}

func (sw* S3Walker) onDelete() {
  opts := minio.RemoveObjectOptions {}
  err := sw.mc.RemoveObject(context.Background(), sw.bucket, sw.pathPrefix, opts)
  if err != nil {
    log.Fatal(err)
  }
  log.Println(sw.pathPrefix)
}
