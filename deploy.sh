#!/bin/bash

servers=("34.212.24.196" "34.213.163.202" "34.216.181.5" "34.221.98.179" "54.201.223.206")


for element in ${servers[@]}
do
    cat remote.sh | ssh -i ~/.ssh/cse223B-G12.pem ec2-user@$element /bin/bash
#     ssh -i ~/.ssh/cse223B-G12.pem ec2-user@$element
done
