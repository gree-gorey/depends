# depends
Ensuring dependency between Kubernetes pods

```console
$ docker run greegorey/depends:1.0.0 --help
Usage of ./depends:
  -interval int
    	Interval in seconds (default 2)
  -services string
    	Services to wait for, separated by comma (default "redis:1")
```
