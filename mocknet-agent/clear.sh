#!/bin/bash

service vpp restart
echo "vpp service restarted"
sudo vppctl set int state GigabitEthernet0/6/0 up
echo "set interface GigabitEthernet0/6/0 state up"
sudo vppctl set int ip address GigabitEthernet0/6/0 10.3.0.1/24
echo "assign ip address 10.3.0.1/24 to interface GigabitEthernet0/6/0"
