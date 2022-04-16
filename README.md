# tinys3cli
A Tiny S3 Client Application in Golang

# install

```sh
go install github.com/lucidfrontier45/tinys3cli@latest
```

# usage

```
# list objects
tinys3cli list s3://bucket/path

# upload objects
tinys3cli put [-j n_jobs] localfile1 [localfile2] ... s3://bucket/path

# download objects (recursively)
tinys3cli get [-r] [-j n_jobs] s3:/bucket/path localpath
```