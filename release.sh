#!/bin/bash
set -e
if [ -z "$1" ]
  then
    echo "Need a tag version. Ex: v1"
    exit 1
fi
TAG=$1
git tag -a ${TAG} -m "${TAG}"
echo "Tagged: ${TAG}"
GIT_SHA=$(git rev-parse --short HEAD)
# TAG=$(git describe --abbrev=0 --tags)
rm -rf bin/
echo "--> Building..."
FILE_NAME="jabil-pill-checker-${TAG}.exe"
GOOS=windows GOARCH=386 go build -o bin/${FILE_NAME} -ldflags="-X main.Commit=${GIT_SHA} -X main.Tag=${TAG}"
GOOS=darwin GOARCH=amd64 go build -o bin/jabil-pill-checker-${TAG} -ldflags="-X main.Commit=${GIT_SHA} -X main.Tag=${TAG}"

aws s3 cp bin/${FILE_NAME} s3://hello-jabil/pill-checker/
echo "Uploaded: s3://hello-jabil/pill-checker/${FILE_NAME}"
echo "The file ${FILE_NAME} is not public. Upload to Arena for ECO."