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
tinys3cli put localfile1 [localfile2] ... s3://bucket/path

# download objects (recursively)
tinys3cli get [-r] s3:/bucket/path localpath
```