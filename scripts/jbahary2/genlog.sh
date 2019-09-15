# Arg 1 = name of file
# Arg 2 = number of quotes
if [ -f $1 ]; then
    rm $1
    touch $1
else
    touch $1 
fi

i=0
randline=$(shuf -i 1-$2 -n 1)
randword=$(shuf -i 1-10 -n 1)

while [ $i -lt $2 ]; do 
    if [ $randline -eq $i ]; then
        r='p'
        sed -n "$randword$r" randwords.txt >> $1
    fi
    curl -X GET 'https://geek-jokes.sameerkumar.website/api' >> $1 
    i=$((i+1))
done
