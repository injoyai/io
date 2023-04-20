#-ldflags="-w -s"
#-ldflags="-H windowsgui"
#-ldflags="-X "

name="portProxyClient"

GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-w -s" -o ./$name

sleep 5
