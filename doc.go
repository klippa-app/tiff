// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package tiff implements a TIFF image decoder and encoder.

The TIFF specification is at http://partners.adobe.com/public/developer/en/tiff/TIFF6.pdf

Classic TIFF Structure:

	+------------------------------------------------------------------------------+
	|                          Classic TIFF Structure                              |
	|  IFH                                                                         |
	| +-----------------+                                                          |
	| | II/MM ([2]byte) |                                                          |
	| +-----------------+                                                          |
	| | 42     (uint16) |      IFD                                                 |
	| +-----------------+     +------------------+                                 |
	| | Offset (uint32) |---->| Num     (uint16) |                                 |
	| +-----------------+     +------------------+                                 |
	|                         | Entry ([12]byte) |                                 |
	|                         +------------------+                                 |
	|                         | Entry ([12]byte) |                                 |
	|                         +------------------+                                 |
	|                         | ...              |      IFD                        |
	|                         +------------------+    +------------------+         |
	|     IFD Entry           | Offset  (uint32) |--->| Num     (uint16) |         |
	|    +-----------------+  +------------------+    +------------------+         |
	|    | Tag    (uint16) |                          | Entry ([12]byte) |         |
	|    +-----------------+                          +------------------+         |
	|    | Type   (uint16) |<-------------------------| Entry ([12]byte) |         |
	|    +-----------------+                          +------------------+         |
	|    | Count  (uint32) |                          | ...              |         |
	|    +-----------------+                          +------------------+         |
	|    | Offset (uint32) |---> Value                | Offset  (uint32) |--->NULL |
	|    +-----------------+                          +------------------+         |
	|                                                                              |
	+------------------------------------------------------------------------------+

Big TIFF Structure:

	+------------------------------------------------------------------------------+
	|                          Big TIFF Structure                                  |
	|  IFH                                                                         |
	| +-----------------+                                                          |
	| | II/MM ([2byte]) |                                                          |
	| +-----------------+                                                          |
	| | 43     (uint16) |                                                          |
	| +-----------------+                                                          |
	| | 8      (uint16) |                                                          |
	| +-----------------+                                                          |
	| | 0      (uint16) |      IFD                                                 |
	| +-----------------+     +------------------+                                 |
	| | Offset (uint64) |---->| Num     (uint64) |                                 |
	| +-----------------+     +------------------+                                 |
	|                         | Entry ([20]byte) |                                 |
	|                         +------------------+                                 |
	|                         | Entry ([20]byte) |                                 |
	|                         +------------------+                                 |
	|                         | ...              |      IFD                        |
	|                         +------------------+    +------------------+         |
	|     IFD Entry           | Offset  (uint64) |--->| Num     (uint64) |         |
	|    +-----------------+  +------------------+    +------------------+         |
	|    | Tag    (uint16) |                          | Entry ([12]byte) |         |
	|    +-----------------+                          +------------------+         |
	|    | Type   (uint16) |<-------------------------| Entry ([12]byte) |         |
	|    +-----------------+                          +------------------+         |
	|    | Count  (uint64) |                          | ...              |         |
	|    +-----------------+                          +------------------+         |
	|    | Offset (uint64) |---> Value                | Offset  (uint64) |--->NULL |
	|    +-----------------+                          +------------------+         |
	|                                                                              |
	+------------------------------------------------------------------------------+

Report bugs to <chaishushan@gmail.com>.

Thanks!
*/
package tiff
