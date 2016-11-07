#! /usr/bin/env bash
#echo "$#"
#echo "$1"
#echo "$2"
if [ "$#" -eq 0 ]; then
  echo "version [old ver] [new ver]"
  exit 1
fi

if [ "$#" -eq 1 ]; then
  tmp=`echo $1 | sed 's/\\./\\\\./g'`
  #echo $tmp
  grep -R --exclude-dir vendor --exclude-dir .git --exclude-dir mesos --exclude scaleio-executor --exclude scaleio-scheduler --exclude build.sh $tmp ./
else
  tmp=`echo $1 | sed 's/\\./\\\\\\./g'`
  #echo $tmp
  find . -type f -name '*.json' -exec sed -i '' 's/'"$tmp"'/'"$2"'/g' {} +
  find . -type f -name '*.md' -exec sed -i '' 's/'"$tmp"'/'"$2"'/g' {} +
  find . -type f -name '*.go' -exec sed -i '' 's/'"$tmp"'/'"$2"'/g' {} +
  find . -type f -name 'VERSION' -exec sed -i '' 's/'"$tmp"'/'"$2"'/g' {} +
fi
