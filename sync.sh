#!/bin/sh
limit=10
count=0
while [ ${count} -lt ${limit} ]
do
  rsync=`rsync -av --delete ${1} ${2} | sed -n 2p`
  if [ "${rsync}" = "" ]; then
    break
  fi
  count=`expr $count + 1`
done
