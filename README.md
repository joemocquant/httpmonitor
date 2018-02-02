# HTTP log monitoring console program #

Simple console program that monitors HTTP traffic
<br><br>

## Features: ##


* Consume an actively written-to w3c-formatted HTTP access log (https://en.wikipedia.org/wiki/Common_Log_Format).
* Display in the console at regular intervals the sections of the web site with the most hits and metrics on the traffic as a whole.
* Sliding window generating real time alerts for high traffic and traffic recovery thresholds.
<br>

Metrics: 

* Requests: number of request for the period
* Errors: number of errors (http status code >= 400) for the period
* Traffic: number of bytes downloaded for the period
* Unique visitors: number of differents ips for the period
* Avg page views per visitor: number of requests in average per visitor for the period
<br><br>

## Design: ##


* Read accesses are done at specific interval (every readFrequency)
* Data are retrieved into different buffers which are processed concurrently
* The number of read accesses at each interval is bound by the size of the buffer and the amount of logs available to be retrieved
* Data are mapped to a data structure (entryQueue) on which computation is done
* Memory pools exists for entries and buffers ensuring stability regarding memory consumption over time (versus relying on garbage collection)
<br><br>

If logs are appended with a latency greater than the latency to fetch entries from the io.Reader, it results in inacurate metrics/alerts. A "delay" parameter chosen accordingly to the context solve the problem.

<br><br>
Timestamp consideration:

It is assumed that timestamps match the request completion time. Therefore all terminated requests happened for the matching period are retrieved (if the delay parameter is correctly chosen).
<br><br>

## Performance: ##


* This configuration (3 GHz Intel Core i7 / 16 GB 1600 MHz DDR3) can easily handle 100k+ log lines per second.
* The parameters need to be adjusted to the specific context and scale.
<br><br>

## Test and Run: ##


To test:
<br>
`make test && make benchmark`

To run:
<br>
`make build && make run`

