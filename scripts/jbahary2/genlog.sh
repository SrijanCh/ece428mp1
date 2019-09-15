# Arg 1 = name of file
# Arg 2 = number of quotes
if [ -f $1 ]; then
    rm $1
fi

i=0

while [ $i -lt $2 ]; do 
    curl -X GET 'https://geek-jokes.sameerkumar.website/api' >> $1 
    i=$((i+1))
done
