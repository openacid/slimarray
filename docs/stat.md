| Data size | Data Set                | gzip size | slimarry size | avg size   | ratio |
| --:       | :--                     | --:       | :--           | --:        | --:   |
| 1,000     | rand u32: [0, 1000]     | x         | 824 byte      | 6 bit/elt  | 18%   |
| 1,000,000 | rand u32: [0, 1000,000] | x         | 702 KB        | 5 bit/elt  | 15%   |
| 1,000,000 | IPv4 DB                 | 2 MB      | 2 MB          | 16 bit/elt | 50%   |
| 600       | [slim][] star count     | 602 byte  | 832 byte      | 10 bit/elt | 26%   |

[slim]: https://github.com/openacid/slim
