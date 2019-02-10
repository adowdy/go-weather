
#!/bin/bash
# setup environment for go to cross compile to arm, n build

set -e
set -x

env GOOS=linux GOARCH=arm GOARM=5 go build
