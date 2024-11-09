# One Billion Row Challenge in Go

1 Billion Row Challenge consists in computing the min, max, and average of 1 billion measurements.
The measurements represents temperatures per weather station.

More details can be found here : https://1brc.dev/

As part of my learning of Go, this was a good opportunity for me to write Go code and learn. 

# Rules 

* No external library dependencies may be used. That means no lodash, no numpy, no Boost, no nothing. You're limited to the standard library of your language.
* Implementations must be provided as a single source file. Try to keep it relatively short; don't copy-paste a library into your solution as a cheat.
* The computation must happen at application runtime; you cannot process the measurements file at build time

* Input value ranges are as follows:

  * Station name: non null UTF-8 string of min length 1 character and max length 100 bytes (i.e. this could be 100 one-byte characters, or 50 two-byte characters, etc.)
  * Temperature value: non null double between -99.9 (inclusive) and 99.9 (inclusive), always with one fractional digit

* There is a maximum of 10,000 unique station names.

* Implementations must not rely on specifics of a given data set. Any valid station name as per the constraints above and any data distribution (number of measurements per station) must be supported.



## Prerequisites

Go 1.23

Instructions to generate the data can be found here : https://github.com/gunnarmorling/1brc?tab=readme-ov-file#running-the-challenge

## Run 
In the module directory : 
```
go run .
```

## Improvements and results

|Improvement|Duration|
|-|-|
|Naive read and map|250s|
|Add concurrency|135s|
|Removal of Debug logs in critical code|35s|
|Parse temperature as integer|24s|
|Parse temperature manually|19s|
|Use int32 hash instead of string as map key|12s|
|Use of NumCPU CPUs|10s|

Over 50 runs with last version, ran with an average of 10.2s

## CPU Profiling

To see what part of code should be improved, I used `pprof` to profile my code. 

In main.go, the following code can be uncommented : 

```
f, err := os.Create("profiles/new_cpu_profile.prof")
if err != nil {
    panic(err)
}
defer f.Close()

if err := pprof.StartCPUProfile(f); err != nil {
    panic(err)
}
defer pprof.StopCPUProfile()
```

It creates a `.prof` file that can then be opened with `go tool pprof -http 127.0.0.1:8080 profiles/new_cpu_profile.prof`. 
It allows to see the directed graph or the flame graph, and determine what takes time during processing. 

## Further improvements
Based on the current CPU profiling graph, improvements efforts could be focused on :
* Map access which is now the main bottleneck.
* Computation of the hash