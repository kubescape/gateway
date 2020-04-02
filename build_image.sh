#! /bin/bash
if [ -z "${PROJECT_NAME}" ]; then 
    echo "Please set PROJECT_NAME variable"
    exit 1
fi
echo "job name: "$JOB_NAME
if [[ ! -z "${JOB_NAME}" ]]; then 
    mkdir $GOPATH/src
    echo creating symlink
    ln -s $GOPATH $GOPATH/src/$PROJECT_NAME
    cd $GOPATH/src/$PROJECT_NAME
    echo "pwd: "
    pwd
    ls
    go get ./... 
    go get github.com/tebeka/go2xunit
fi
echo "GOPATH= "$GOPATH
CC=$($(which musl-gcc) go build -o dist/$PROJECT_NAME --ldflags '-w -linkmode external -extldflags "-static"' *.go)
CC=$?
echo "alpine build result:" $CC
if [[ $CC != "0" ]]; then
    exit 1
fi
rm -rf *tests_xunit.xml
rm -rf *_tests_go.txt
for path in ./*; do
    [ -d "${path}" ] || continue # if not a directory, skip    
    dirname="$(basename "${path}")"
    if [[ $dirname == *"jenkinstools"* ]] || [[ $dirname == *"dist"* ]] || [[ $dirname == "src" ]] || [[ $dirname == "pkg" ]] || [[ $dirname == "bin" ]] || [[ $dirname == *"component_test"* ]]; then
        continue
    fi
    echo testing "${dirname}"
    go test -v ./$dirname  > ${dirname}_tests_go.txt
    cat  ${dirname}_tests_go.txt | $GOPATH/bin/go2xunit >${dirname}_tests_xunit.xml
done
TESTS_FAILED=$(find -type f -name "*.txt" -exec grep -l 'FAIL' {} +)
if [ -z "${TESTS_FAILED}" ]; then 
    echo "<---------------GOLANG Tests passed:---------------------->"
    echo $TESTS_FAILED
else
    echo "GOLANG Failed tests: $TESTS_FAILED"
    exit 1
fi
