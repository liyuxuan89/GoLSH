# GoLSH
local sensitive hashing for nearest neighbor search

## A simple frontend for color based image retrieval

https://github.com/liyuxuan89/Image-Search-Frontend

<img src="./images/demo.gif">

## Benchmark

test with https://github.com/tsenart/vegeta

100w rows of data, feature dimension 32
```shell
./vegeta -cpus 4  attack -targets target.txt -max-workers 50 -body body.json -timeout=20s -rate 0  -duration=10s  | ./vegeta report -output result.bin
```
MySQL
```
Requests      [total, rate, throughput]         2629, 262.89, 259.70
Duration      [total, attack, wait]             10.123s, 10s, 122.937ms
Latencies     [min, mean, 50, 90, 95, 99, max]  644.813µs, 191.451ms, 168.535ms, 365.615ms, 435.246ms, 646.962ms, 931.302ms
Bytes In      [total, mean]                     44693, 17.00
Bytes Out     [total, mean]                     10516, 4.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:2629  
Error Set:
```
Redis
```
Requests      [total, rate, throughput]         11523, 1152.28, 1151.36
Duration      [total, attack, wait]             10.008s, 10s, 8.021ms
Latencies     [min, mean, 50, 90, 95, 99, max]  403.683µs, 43.302ms, 38.028ms, 99.614ms, 118.35ms, 151.393ms, 249.854ms
Bytes In      [total, mean]                     195891, 17.00
Bytes Out     [total, mean]                     46092, 4.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:11523  
Error Set:
```
