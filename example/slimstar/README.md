# Compressed list of repo stars stats

In this example we use SlimArray to store daily star count of repo [slim](https://github.com/openacid/slim) :

[![Stargazers over time](https://starchart.cc/openacid/slim.svg)](https://starchart.cc/openacid/slim)

- The original star count file is 3129 bytes [./slim-stars-2019-02-02-to-2020-11-14.txt](./slim-stars-2019-02-02-to-2020-11-14.txt).
- The gzip-ed daily star count is 602 bytes [./slim-stars-2019-02-02-to-2020-11-14.txt.gz](./slim-stars-2019-02-02-to-2020-11-14.txt.gz).
- The slimarray with these data is 832 bytes(run `go run slimstar.go`).
    ```
    map[bits/elt:10 elt_width:6 mem_elts:424 mem_total:832 n:652 seg_cnt:1 span_cnt:7 spans/seg:6]
    ```

It shows that SlimArray memory cost is quite close to size of gzip-ed data.
But it provides instant access to the data without the need to decompress!


# Usage

```sh
go run slimstar.go
```

