#!/usr/bin/env bash

while read line
do
	echo -n "$line" >> /var/tmp/facts
done < "${1:-/proc/${$}/fd/0}"
