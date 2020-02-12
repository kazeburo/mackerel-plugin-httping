# mackerel-plugin-httping

Custom mackerel plugin for measuring http/https latency

mackerel-plugin-httping does

- preflight request for keepalive and dnscache before measure latency
- request to the url 10 times sequentially
- use first byte timing as latency
- calculate min, max, 90%tile and average

# Usage

```
% ./mackerel-plugin-httping -h
Usage:
  mackerel-plugin-httping [OPTIONS]

Application Options:
      --url=               URL to ping
      --timeout=           timeout millisec per ping (default: 5000)
      --interval=          sleep millisec after every ping (default: 10)
      --count=             Count Sending ping (default: 10)
      --key-prefix=        Metric key prefix
      --disable-keepalive  disable keepalive
  -v, --version            Show version

Help Options:
  -h, --help               Show this help message
```

## output sample

```
% ./mackerel-plugin-httping --key-prefix example  --url http://example.com/
httping.example_rtt_count.success       10.000000       1581488858
httping.example_rtt_count.error 0.000000        1581488858
httping.example_rtt_ms.max      111.961772      1581488858
httping.example_rtt_ms.min      108.004440      1581488858
httping.example_rtt_ms.average  109.899642      1581488858
httping.example_rtt_ms.90_percentile    110.906144      1581488858
```
