package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/s3utils"
)

func NewCustomHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          1024,
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		DisableCompression:    true,
	}
}

type S3Static struct {
	bucket     string
	bucketPath string
	client     *minio.Client
}

type ObjectMeta struct {
	Objects []minio.ObjectInfo
}

func NewS3Static(accessKey, secretKey, bucket, bucketPath string, endpoint *url.URL) *S3Static {

	client, err := minio.New(endpoint.Host, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       endpoint.Scheme == "https",
		Region:       s3utils.GetRegionFromURL(*endpoint),
		BucketLookup: minio.BucketLookupAuto,
		Transport:    NewCustomHTTPTransport(),
	})

	if err != nil {
		log.Fatalln(err)
	}

	if bucketPath = path.Clean(bucketPath); bucketPath == "." || bucketPath == pathSeparator {
		bucketPath = ""
	} else {
		bucketPath = strings.TrimPrefix(bucketPath, pathSeparator) + "/"
	}

	// bucketPath will be "" or "path/"

	return &S3Static{
		bucket:     bucket,
		bucketPath: bucketPath,
		client:     client,
	}

}

func (s *S3Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	URI, err := url.QueryUnescape(r.RequestURI)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ReqDir := strings.HasSuffix(URI, pathSeparator)
	ReqPath := path.Clean(URI) //ReqPath should be "" or "path" or "path/"

	if ReqPath == pathSeparator {
		ReqPath = ""
	} else if ReqDir {
		ReqPath = strings.TrimPrefix(ReqPath, pathSeparator) + pathSeparator
	} else {
		ReqPath = strings.TrimPrefix(ReqPath, pathSeparator)
	}

	FullPath := s.bucketPath + ReqPath

	meta, exist, file := s.GetObjectMeta(r.Context(), FullPath, ReqDir)

	// 404

	if !exist {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 302

	if !ReqDir && !file {
		w.Header().Add("Location", ReqPath+pathSeparator)
		w.WriteHeader(http.StatusFound)
		return
	}

	if !file {

		// folder

		w.Header().Set("Content-Type", `text/html; charset=utf-8`)
		fmt.Fprintf(w, "<pre>\n")

		for _, object := range meta.Objects {
			key := strings.Replace(object.Key, FullPath, "", 1)
			fmt.Fprintf(w, `<a href="%s">%s</a>`+"\n", key, key)
		}

		fmt.Fprintf(w, "</pre>")

	} else {

		// file

		object, err := s.client.GetObject(r.Context(), s.bucket, FullPath, minio.GetObjectOptions{})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		stat, err := object.Stat()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		http.ServeContent(w, r, stat.Key, stat.LastModified, object)
	}
}

func (s *S3Static) GetObjectMeta(ctx context.Context, fullPath string, ReqDir bool) (meta *ObjectMeta, exist, file bool) {

	meta = &ObjectMeta{}

	for object := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    fullPath,
		Recursive: false,
	}) {
		meta.Objects = append(meta.Objects, object)
	}

	if length := len(meta.Objects); length >= 1 {

		if meta.Objects[0].Key == fullPath && length == 1 {
			return meta, true, true
		}

		return meta, true, false
	}

	return meta, false, false
}
