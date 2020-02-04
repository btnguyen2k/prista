`prista` is a service that client can throw logs in. Logs are then either stored or forwarded to another sink.

GitHub repository: https://github.com/btnguyen2k/prista

## Quick Start

```shell script
% docker run -d -p 8070:8070/udp -p 8080:8080 -p 8090:8090 bbtnguyen2k/prista:<tag>
```

- Port 8070: UDP server
- Port 8080: HTTP server
- Port 8090: gRPC server

# Environment Variables

| Env               | Default Value    | Description |
|-------------------|:----------------:|-------------|
| TIMEZONE          | Asia/Ho_Chi_Minh | Timezone to use in date/time-related operations. |
| HTTP_LISTEN_ADDR  | 0.0.0.0          | HTTP server's listen address. |
| HTTP_LISTEN_PORT  | 8080             | HTTP server's listen port. Set to 0 to disable HTTP server. |
| GRPC_LISTEN_ADDR  | 0.0.0.0          | gRPC server's listen address. |
| GRPC_LISTEN_PORT  | 8090             | gRPC server's listen port. Set to 0 to disable gRPC server. |
| UDP_LISTEN_ADDR   | 0.0.0.0          | UDP server's listen address. |
| UDP_LISTEN_PORT   | 8070             | UDP server's listen port. Set to 0 to disable gRPC server. |
| UDP_THREADS       | 4                | (since v0.1.2) Number of threads to handle messages sent via UDP. |
| MAX_REQUEST_SIZE  | 4kB              | Max request size (imply max log entry size). Format: absolute number means `size in bytes` or number+suffix, see https://github.com/lightbend/config/blob/master/HOCON.md#size-in-bytes-format |
| REQUEST_TIMEOUT   | 10s              | Timeout to read request data. Format: absolute number means `time in milliseconds` or number+suffix, see https://github.com/lightbend/config/blob/master/HOCON.md#duration-format |
| TEMP_DIR          | ./temp           | "temp" directory to buffer incoming log messages. |
| MAX_WRITE_THREADS | 128              | (since v0.1.2) Max number of concurrent log writes. |
