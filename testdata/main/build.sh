#-ldflags="-w -s"
#-ldflags="-H windowsgui"
#-ldflags="-X "

name="portProxyServer"

GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name

sleep 5