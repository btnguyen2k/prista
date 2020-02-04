# prista Release Notes

## 2020-02-04 - v0.1.2

- Improve UDP server & new config `server.udp.num_threads`
- Improve log writing & new config `max_write_threads`


## 2020-02-02 - v0.1.1

- Add funnction `ILogWriter.Info() map[string]interface{}`
- Add new log writer config `retry_seconds`
- New `forward` log writer that forwards log entries to another `prista` instance


## 2020-01-28 - v0.1.0

First release:

- Collect logs via HTTP, gRPC and HTTP
- Write logs to file on disk, support file rotation
