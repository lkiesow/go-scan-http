go-scan-http
============

> Fast http network scanner

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.


Usage
-----

```sh
go-scan-http addr-range | [b1 ... b4]  [ports ...]

addr-range
  Address range specification as single string in the form
  b1.b2.b3.b4/mask. E.g. `192.168.1.0/24`
b[1-4]
  Specification for a byte range to scan.
  E.g. `10` or `1-254` or `100-150`.
  Using the special value `*` is equivalent to `1-254`.
ports
  List of ports to scan.
  This defaults to 80.

Example
  Scan a 192.168.1.0/24 network for ports 80 and 8080.
  All these forms are equivalent.
    go-scan-http 192.168.1.0/24  80 8080
    go-scan-http 192 168 1 1-254 80 8080
    go-scan-http 192 168 1 *     80 8080
```
