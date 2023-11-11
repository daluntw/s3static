package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	pathSeparator = "/"
)

var (
	initEndpoint   string
	initAccessKey  string
	initSecretKey  string
	initAddress    string
	initBucket     string
	initBucketPath string
)

func defaultEnvString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func init() {
	flag.StringVar(&initEndpoint, "endpoint", defaultEnvString("S3_ENDPOINT", ""), "AWS S3 compatible server endpoint")
	flag.StringVar(&initBucket, "bucket", defaultEnvString("S3_BUCKET", ""), "bucket name with static files")
	flag.StringVar(&initBucketPath, "bucketPath", defaultEnvString("S3_BUCKET_PATH", ""), "bucket path to serve static files from")
	flag.StringVar(&initAccessKey, "accessKey", defaultEnvString("S3_ACCESS_KEY", ""), "access key for server")
	flag.StringVar(&initSecretKey, "secretKey", defaultEnvString("S3_SECRET_KEY", ""), "secret key for server")
	flag.StringVar(&initAddress, "address", defaultEnvString("S3_ADDRESS", "127.0.0.1:8080"), "bind to a specific ADDRESS:PORT, ADDRESS can be an IP or hostname")
}

func main() {

	flag.Parse()

	if strings.TrimSpace(initBucket) == "" {
		log.Fatalln(`Bucket name cannot be empty, please provide '-bucket "{mybucket}"'`)
	}

	u, err := url.Parse(initEndpoint)
	if err != nil {
		log.Fatalln(err)
	}

	initBucketPath = strings.TrimPrefix(initBucketPath, pathSeparator)

	log.Fatal(http.ListenAndServe(initAddress, NewS3Static(
		initAccessKey,
		initSecretKey,
		initBucket,
		initBucketPath,
		u,
	)))
}
