#!/usr/bin/env bash

while true
do 
	OUTPUT="$(service vpp status | grep active)"
	if [ -z "$OUTPUT" ];then
		echo "waiting for vpp service to restart"
		sleep 1
	else
		break 1
	fi
done