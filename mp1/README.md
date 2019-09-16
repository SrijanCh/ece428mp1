# ece428mp1
ece 428 mp 1
In order to set up, begin an external client, set up a Go workspace, and then move to where source should be and pull this repository into a file called source there. Our repository IS the src file due to how stupidly Go treats libraries. Make sure to set up GOPATH! Instructions on this are in Go documentation and not our focus.

We have provided shell scripts to facilitated the setup. Move to GO_WORKSPACE/src/scripts/srijanc2 in a terminal. Run the scripts in this order:

sh gostart.sh killhost
sh gostart.sh update
sh gostart.sh start

In order, these scripts kill all the hosts if any were left, pull the latest code from git, and finally goes through each server, builds the basic client and host, and runs the server in the background on each machine.

In order to run a grep on the given logs, we must move to GO_WORKSPACE/src/mp1, and call the client with the flags and strings to search that we would usually pass grep. We don't have to specify file name; since this is a distributed log query, the logs are found for us. These logs are the vmlogs passed to us by the class; they have been put in the machines already. Here's an example call:

./client [flags] [arg]

And this should perform the grep over all 10 nodes. We've also included a loggrep for the same things but machine.i.log files sitting in home instead. We have to build this manually, but here otherwise it's the same grep, different file.

go build loggrep.go
./logrep [flags] [arg]

Finally, we've included a reg_grep. This one works like regular grep, down to allowing passing the filename. This is not too useful, however, because we need to find the exact path on every machine in the cluster to pass as args. In other words, it'll run the same grep on every machine.

go build reg_grep.go
./reg_grep [flags] [arg] [filename] 

The latter two are technically not required by the MP and were added for fun. One important thing to remember is that all the logs, in the form of either vm#.log or machine.i.log, must be pre set up on each machine's home beforehand. We used scripts to do so, but we leave the user to use any method they wish. Finally, to kill the current machine, just kill the host process on it, and it will be dead to the cluster:

killall host