i=1
while [ $i -lt 11 ]; do
    num=$i
    if [ $i -lt 10 ]; then
       num="0$i" 
    fi
    echo "Sshing to fa19-cs425-g77-$num.cs.illinois.edu"
    sshpass -f ~/pass.txt ssh srijanc2@"fa19-cs425-g77-$num.cs.illinois.edu" "cd; 
    export GOPATH=/home/srijanc2/go_work/
    cd /home/srijanc2/go_work/src/mp3/
    nohup go run membership.go -logfile=vmlog.txt > /dev/null 2>&1 &
    "
    i=$((i+1))
done 
