windres resources.rc -o resources.syso
go build -ldflags "-extld=gcc -extldflags=resources.syso -H=windowsgui" main.go