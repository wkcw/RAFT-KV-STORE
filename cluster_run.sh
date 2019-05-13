#!/bin/bash

servers=("52.10.200.237" "54.214.177.198" "34.219.38.122" "18.236.134.213" "52.40.150.174")
node_num=(6 6 6 6 7)


if [ $1 == "run" ]; then
    for (( i=0; i<${#servers[@]}; i++ ))
    do
        node_id=0
        for (( j=0; j<node_num[i]; j++ ))
        do
            cat run.sh | sed -e 's/#node_num/node'${node_id}'/g' | ssh -i ~/.ssh/cse223B-G12.pem ec2-user@${servers[i]} /bin/bash
            node_id+=1
        done
    done

elif [ $1 == "deploy" ]; then
    for (( i=0; i<${#servers[@]}; i++ ))
    do
#      echo ${servers[i]}
#       cat deploy.sh | ssh -i ~/.ssh/cse223B-G12.pem ec2-user@${servers[i]} /bin/bash

     ssh -i ~/.ssh/cse223B-G12.pem ec2-user@${servers[i]}
    done

else
    echo "invalid command line parameters"

fi