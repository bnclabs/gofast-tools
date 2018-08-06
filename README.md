README
======

[![IRC #bnc](https://www.irccloud.com/invite-svg?channel=%23bnc&amp;hostname=irc.mozilla.org&amp;port=6667)](https://www.irccloud.com/invite?channel=%23bnc&amp;hostname=irc.mozilla.org&amp;port=6667)

Collection of tools for Gofast quality control.

* `verify/` sub-package is a standalone program to thrash all gofast
  server and client APIs with all types of messages.
* `plot/` directory has Javascript program to plot gofast statistics
  in a graph, can be useful for performance measurements.

```go
$ make check
```
