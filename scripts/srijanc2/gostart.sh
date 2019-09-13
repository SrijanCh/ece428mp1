i=1
while [ $i -lt 11 ]; do
    num=$i
    if [ $i -lt 10 ]; then
       num="0$i" 
    fi
    echo "Sshing to fa19-cs425-g77-$num.cs.illinois.edu"
    sshpass -f pass.txt ssh srijanc2@"fa19-cs425-g77-$num.cs.illinois.edu" "cd; 
        mkdir go_work; 
        echo 'GOPATH=$HOME/go_work/src' >> .bashrc; 
        cd go_work;
        git clone https://github.com/SrijanCh/ece428mp1 src;
        "
    i=$((i+1))
done 
