#!/bin/bash

speedUp=$1
speedDown=$2

# Удаление очередей
/sbin/tc qdisc del dev eth1 ingress
/sbin/tc qdisc del dev eth1 root handle 1:

# Ограничение скорости отдачи
/sbin/tc qdisc add dev eth1 root handle 1: htb default 10 r2q 1
/sbin/tc class add dev eth1 parent 1: classid 1:10 htb rate ${speedUp}kbit quantum 8000 burst 8k

# Ограничение скорости загрузки
/sbin/tc qdisc add dev eth1 handle ffff: ingress
/sbin/tc filter add dev eth1 parent ffff: protocol ip prio 50 u32 match ip src 0.0.0.0/0 police rate ${speedDown}kbit burst 12k drop flowid :1