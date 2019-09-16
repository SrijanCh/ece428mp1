import random
import os
import socket

NUMLINES = 50
MAXWORDSONLINE = 10
MINWORDSONLINE = 1
HOSTNAMEREPEATINTERVAL = 5
LOGFILEOUTPUTNAME = "machine.i.log"

outputstring = ""
with open('engmix.txt', 'r') as myfile:
    filecontent = myfile.read()
    words = filecontent.split('\n')
    linenum = 1
    while linenum <= NUMLINES:
        # create random number of words on the line 
        numwords = random.randint(MINWORDSONLINE, MAXWORDSONLINE + 1)
        wordnum = 1
        if linenum % HOSTNAMEREPEATINTERVAL == 0:
            outputstring += socket.gethostname()
        else:
            while wordnum <= numwords:
                randwordnum = random.randint(0, len(words))
                outputstring += words[randwordnum] + " "
                wordnum += 1

        outputstring += '\n'
        linenum += 1

with open(LOGFILEOUTPUTNAME, 'w') as myfile:
    myfile.write(outputstring)



