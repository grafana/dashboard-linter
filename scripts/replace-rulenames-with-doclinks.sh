#! /usr/bin/env bash

for rulefile in ./docs/rules/*.md
do 
  rulename=$(basename $rulefile .md)
  for docfile in $(find ./docs -regex ".*\.md\|.*_intermediate/.*\.txt" -print)
  do
   sed -i".bak" "s,\`${rulename}\`,\[${rulename}\]\(./rules/${rulename}.md\),g" "${docfile}"
   rm "${docfile}.bak"
  done
done

