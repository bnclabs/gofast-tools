* golang 1.6 has default support for http 2.0. benchmark gofast with
  http 2.0.
* try gofast on raspberry-pi.
* test case with 1-byte packet.
* test case with 0-byte packet.
* find a way to add SendHeartbeat() in verify/{client.go,server.go}
* $ client :9900 stream # fails for some reason. make it more resilient.
