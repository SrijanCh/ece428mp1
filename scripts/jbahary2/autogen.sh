i=1
if [ -f iptables.txt ]; then
    rm iptables.txt
fi
touch iptables.txt
while [ $i -lt 11 ]; do 
    num=$i
    if [ $i -ne 10 ]; then
        num="0$i"
    fi
    echo "fa19-cs425-g77-$num.cs.illinois.edu" >> iptables.txt
    i=$((i+1))
done
