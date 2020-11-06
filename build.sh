CGO_CFLAGS="-I$ACLDIR/inc"
CGO_LDFLAGS="-L$ACLDIR/lib"
export CGO_CFLAGS CGO_LDFLAGS
go build --tags adalnk .

