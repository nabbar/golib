#!/usr/bin/env bash

[ ! -e vendor.tar.gz ] || rm -v vendor.tar.gz
[ ! -e vendor.tar.bz2 ] || rm -v vendor.tar.bz2
[ ! -e vendor.zip ] || rm -v vendor.zip
[ ! -e ph-000.tar ] || rm -v ph-00*.tar
[ ! -e test-archive ] || rm -v test-archive

tar cf ph-000.tar vendor/ && \
cp -v ph-000.tar ph-001.tar && \
cp -v ph-000.tar ph-002.tar && \
cp -v ph-000.tar ph-003.tar && \
cp -v ph-000.tar ph-004.tar && \
cp -v ph-000.tar ph-005.tar && \
cp -v ph-000.tar ph-006.tar && \
cp -v ph-000.tar ph-007.tar && \
cp -v ph-000.tar ph-008.tar && \
cp -v ph-000.tar ph-009.tar && \
tar czf vendor.tar.gz vendor/ && \
tar cjf vendor.tar.bz2 ph-00*.tar vendor.tar.gz && \
zip vendor.zip ph-00*.tar vendor.tar.bz2 && \
rm -vf ph-00*.tar vendor.tar.* && \
ls -ahl vendor.zip && \
go build -a -v -o test-archive ./test/test-archive/ && \
echo -e "\n\n\t>>Try to print contents of file github.com/gin-gonic/gin/internal/json/json.go" && \
./test-archive
echo -e "\n\n\t>>from archive zip >> bz2 >> tar >> gz >> tar :" && \
ls -ahl vendor.zip && \
echo -e "\n\n"

[ ! -e vendor.tar.gz ] || rm -v vendor.tar.gz
[ ! -e vendor.tar.bz2 ] || rm -v vendor.tar.bz2
[ ! -e vendor.zip ] || rm -v vendor.zip
[ ! -e ph-000.tar ] || rm -v ph-00*.tar
[ ! -e test-archive ] || rm -v test-archive
