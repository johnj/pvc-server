#!/usr/bin/env bash

while read line
do
	echo -n "$line" >> /var/tmp/reports
done < "${1:-/proc/${$}/fd/0}"
