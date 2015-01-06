sarstats
========

Daemon that takes stats from sar and sends them to statsd every interval


usage
========
```
  -d=":8125": Destination statsd server address
  -i=10s: Interval to send metrics
  -p="sar": Statsd prefix for metrics
```
