#! /bin/bash
ulimit -n 50000
monitorM=`ps -ef | grep nuclei | grep -v grep | wc -l ` 
if [ $monitorM -eq 0 ] 
then
	echo "nuclei is not running, restart nuclei "
	 ./nuclei  >>m.log &
else
	echo "nuclei is running"
fi

