#!/bin/bash

# 创建force_include.go, 目的是强制在vendor中包含需要的css, js等文件，否则引用gonsole的项目可能找不到相关文件

FILENAME=dist/force_include.go
echo "package dist" > $FILENAME
echo  >> $FILENAME
echo "func ForceIncludeFiles() { }" >> $FILENAME

FILENAME=dist/assets/force_include.go
echo "package assets" > $FILENAME
echo  >> $FILENAME
echo "func ForceIncludeFiles() { }" >> $FILENAME

#cp -R statics dist/
cp src/favicon.ico dist/
