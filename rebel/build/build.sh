 #!/bin/bash

build=0
if [ -e ./.buildcount ]; then
    ((build=`cat ./.buildcount`+1))
    echo $build > ./.buildcount
fi

export GOOS="linux"
export GOARCH="amd64"

appname="rebel"
version="1.0.${build}"
filename="${appname}-linux-amd64.out"
doRun=false
isRelease=false

for var in "$@"; do
    if [ "$var" = "--run" ]; then
        doRun=true
    fi
    if [ "$var" = "--release" ]; then
        isRelease=true
    fi
done

if [ $isRelease = true ]; then
    go build \
    -o ./$filename \
    -ldflags "-X main.product_release=true -X main.product_version=${version} -s -w" \
    ./..
else
    go build \
    -o ./$filename \
    -ldflags "-X main.product_release=true -X main.product_version=${version}" \
    ./..
fi

if [ $doRun = true ]; then
    #chmod +x rebel-linux-amd64.out
    ./rebel-linux-amd64.out
fi
