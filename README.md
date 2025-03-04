# go-cache-prog

An implementation of custom Go build cache program.

## go version dependency

go >= 1.24.0

## build

```
$git clone https://github.com/bigwhite/go-cache-prog.git
$make
```

## usage

- build/install go module without cache

```
$GOCACHEPROG="./go-cache-prog --verbose" go install fmt
2025/03/04 10:47:59 Using cache directory: /Users/tonybai/.gocacheprog
2025/03/04 10:47:59 Received request: ID=1, Command=get, ActionID=90c776cb58a3c3a99b5622344df5bc959fd2b90f299b40ae21ec6ccf16c77a23, OutputID=, BodySize=0
2025/03/04 10:47:59 Received request: ID=2, Command=put, ActionID=90c776cb58a3c3a99b5622344df5bc959fd2b90f299b40ae21ec6ccf16c77a23, OutputID=4e67091862cdc5ff3d44d51adaf9f5a3f5e993dcbc0b6aad884d00d929f3f4d3, BodySize=3037
2025/03/04 10:47:59 Put request: ID=2, Actual BodyLen=4055
2025/03/04 10:47:59 Received request: ID=3, Command=get, ActionID=b2d3027bda366ae198f991d65f62b5be25aa7fe41092bb81218ba24363923b69, OutputID=, BodySize=0
2025/03/04 10:47:59 Received request: ID=4, Command=get, ActionID=c48dafcc394ccfed5c334ef2e21ba8b5bd09a883956f17601cf8a3123f8afd2b, OutputID=, BodySize=0
2025/03/04 10:47:59 Received request: ID=5, Command=get, ActionID=b16400d94b83897b0e7a54ee4223208ff85b4926808bcae66e488d2dbab85054, OutputID=, BodySize=0
2025/03/04 10:47:59 Received request: ID=6, Command=get, ActionID=789f5b8e5b2390e56d26ac916b6f082bfb3e807ee34302f8aa0310e6e225ac77, OutputID=, BodySize=0

... ...
2025/03/04 10:48:03 Received request: ID=321, Command=close, ActionID=, OutputID=, BodySize=0
2025/03/04 10:48:03 Gets: 107, GetMiss: 107
```

- build/install go module after cache built

```
$GOCACHEPROG="./go-cache-prog --verbose" go install fmt
2025/03/04 10:50:14 Using cache directory: /Users/tonybai/.gocacheprog
2025/03/04 10:50:14 Received request: ID=1, Command=get, ActionID=90c776cb58a3c3a99b5622344df5bc959fd2b90f299b40ae21ec6ccf16c77a23, OutputID=, BodySize=0
2025/03/04 10:50:14 Received request: ID=2, Command=get, ActionID=c48dafcc394ccfed5c334ef2e21ba8b5bd09a883956f17601cf8a3123f8afd2b, OutputID=, BodySize=0
2025/03/04 10:50:14 Received request: ID=3, Command=get, ActionID=b16400d94b83897b0e7a54ee4223208ff85b4926808bcae66e488d2dbab85054, OutputID=, BodySize=0
2025/03/04 10:50:14 Received request: ID=4, Command=get, ActionID=789f5b8e5b2390e56d26ac916b6f082bfb3e807ee34302f8aa0310e6e225ac77, OutputID=, BodySize=0
2025/03/04 10:50:14 Received request: ID=5, Command=get, ActionID=c6e6427a15f95d70621df48cc68ab039075d66c1087427eb9a04bcf729c5b491, OutputID=, BodySize=0
... ...
2025/03/04 10:50:14 Received request: ID=161, Command=close, ActionID=, OutputID=, BodySize=0
2025/03/04 10:50:14 Gets: 160, GetMiss: 0
```
