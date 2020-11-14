# Compressed list of IPv4

In this example we use SlimArray to store a list of IPv4, which is obtained from
a real IP database.

It shows that SlimArray memory cost is quite close to size of gzip-ed data.
But it provides instant access to the data without the need to decompress!

SlimArray consumes 2.1 MB while the gzip-ed iplist(`./iplist.go.gz`) consumes 2.0 MB:

```
map[bits/elt:18 elt_width:16 mem_elts:1678448 mem_total:2111640 n:902190 seg_cnt:882 span_cnt:13092 spans/seg:14]
```

# Usage

It loads the ip list into SlimArray and print the stat:

```sh
make
```

