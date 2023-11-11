# S3Static

Serve static files from S3 compatible object storage services 

### Usage
```bash
s3static -endpoint "https://play.min.io/" -accessKey "accessKey" -secretKey "secretKey" -bucket "bucket" -bucketPath "public" -address "0.0.0.0:8080" 
```
OR
```bash
S3_ENDPOINT="https://play.min.io/" S3_ACCESS_KEY="accessKey" S3_SECRET_KEY="secretKey" S3_BUCKET="bucket" S3_BUCKET_PATH="public" S3_ADDRESS="0.0.0.0:8080" s3static
```
## Reference

- [s3www](https://github.com/harshavardhana/s3www)