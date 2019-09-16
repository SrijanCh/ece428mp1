host=hostname

go build loggrep.go

python genlogs.py

./loggrep -n $host

python genlogsregexdate.py

./loggrep -n -E "[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}.[0-9]*" 
