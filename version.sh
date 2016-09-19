#! /usr/bin/env bash
#echo "$#"
if [ "$#" -eq 0 ]; then
  echo "version [old ver] [new ver]"
  exit 1
fi

#echo "$1"
#echo "$2"

if [ "$#" -eq 1 ]; then
  grep -R --exclude-dir vendor --exclude-dir .git --exclude-dir mesos --exclude scaleio-executor --exclude scaleio-scheduler --exclude build.sh $1 ./
else
  find . -type f -name '*.json' -exec sed -i '' 's/'"$1"'/'"$2"'/g' {} +
  find . -type f -name '*.md' -exec sed -i '' 's/'"$1"'/'"$2"'/g' {} +
  find . -type f -name '*.go' -exec sed -i '' 's/'"$1"'/'"$2"'/g' {} +
  find . -type f -name 'VERSION' -exec sed -i '' 's/'"$1"'/'"$2"'/g' {} +
fi
