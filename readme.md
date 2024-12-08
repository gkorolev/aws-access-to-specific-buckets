## HowTo

### Install

```bash
brew tap tinygo-org/tools
brew install tinygo

sudo curl https://func-e.io/install.sh | bash
sudo mv ./bin/func-e /usr/local/bin
```

### Test

#### Amazon S3 provides 2 modes to access S3 buckets and objects there:
 - Path-style access: https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html#path-style-access
 - Virtual-hosted-style access: https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html#virtual-hosted-style-access

Run the following to setup http routing, compile wasm and run envoy locally:
```bash
export http_proxy=http://localhost:8080
tinygo build -o main.wasm -scheduler=none -target=wasi main.go
func-e run -c envoy.yaml
```

##### Tests

 - Virtual-hosted-style access:
Should allow access:
```bash
curl  -I amzn-s3-demo-bucket1.s3.us-west-2.amazonaws.com
```
Should block access:
```bash
curl  -I amzn-s3-demo-bucket2.s3.us-west-2.amazonaws.com
```

 - Path-style access:
Should allow access:
```bash
curl  -I s3.us-west-2.amazonaws.com/amzn-s3-demo-bucket1
```
Should block access:
```bash
curl  -I s3.us-west-2.amazonaws.com/amzn-s3-demo-bucket2
```
