#!/bin/bash

set -e

(echo "package mail"

echo "const mailHtml = \`$(cat mail.html)\`"
) | gofmt > static.go.new

mv static.go.new static.go
