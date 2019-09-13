i=1
while [ $i -lt 11 ]; do
    num=$i
    if [ $i -lt 10 ]; then
       num="0$i" 
    fi
    echo "Sshing to fa19-cs425-g77-$num.cs.illinois.edu"
    sshpass -f pass.txt ssh srijanc2@"fa19-cs425-g77-$num.cs.illinois.edu" "cd; 
            cd ~
            touch machine.i.log
            echo -e "$ith machine says hello\n" > machine.i.log
        "
    i=$((i+1))
done 
