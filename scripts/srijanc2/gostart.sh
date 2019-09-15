i=1
while [ $i -lt 11 ]; do
    num=$i
    if [ $i -lt 10 ]; then
       num="0$i" 
    fi
    echo "Sshing to fa19-cs425-g77-$num.cs.illinois.edu"
    sshpass -f pass.txt ssh srijanc2@"fa19-cs425-g77-$num.cs.illinois.edu" "cd; 
        if [ $1 == "setup" ]; then
            mkdir go_work; 
            echo 'GOPATH=$HOME/go_work/src' >> .bashrc; 
            cd go_work;
            git clone https://github.com/SrijanCh/ece428mp1 src;
        elif [ $1 == "update" ]; then
            cd go_work/src
            git pull
        elif [ $1 == "killhost" ]; then
            killall host
        elif [ $1 == "start" ]; then
            cd go_work/src/mp1
            go build client.go
            go build host.go
            # nohup ./host &>/dev/null &
            nohup ./host &>nohup.out &
        else
            echo "Unrecognized command line option"
        fi
        "
    i=$((i+1))
done 
