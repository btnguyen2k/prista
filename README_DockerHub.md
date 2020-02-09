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

**Server configuration environment variables:**

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

**Log writer configuration environment variables:**

| Env                             | Default Value             | Description |
|---------------------------------|:-------------------------:|-------------|
| LOG_DEFAULT_TYPE                | console                   | (*) Log writer type for `default` category. |
| LOG_DEFAULT_CONSOLE_TARGET      | stdout                    | (Config for `console` log writer for `default` category) target to write logs to, either "stdout" or "stderr". |
| LOG_DEFAULT_FILE_ROOT           | ./log/default             | (Config for `file` log writer for `default` category) root directory to store log files. |
| LOG_DEFAULT_FILE_PATTERN        | default.log-20060102_1504 | (**) (Config for `file` log writer for `default` category) name of the log file. |
| LOG_DEFAULT_FILE_TYPE           | json                      | (***) (Config for `file` log writer for `default` category) log content type, either `tsv` or `json`. |
| LOG_DEFAULT_FILE_RETRIES        | 60                        | (-) (Config for `file` log writer for `default` category) retry duration (in seconds). |
| LOG_DEFAULT_FORWARD_DESTINATION |                           | (+) (Config for `forward` log writer for `default` category) destination to forward log to. |
| LOG_DEFAULT_FORWARD_RETRIES     | 180                       | (-) (Config for `forward` log writer for `default` category) retry duration (in seconds). |
| LOG_DEFAULT_FANOUT_TARGETS      |                           | (++) (Config for `fanout` log writer for `default` category) names of other categories to fan-out logs to . |

(-) If log is failed to be written, the write is retrying for (at least) a number of seconds before the log entry is discarded
- value of `0`: no retry
- negative value: retry forever!

(*) Built-in log writer types:
- `console`: (available since `v0.1.4`) write logs to stdout/stderr.
- `file`: write logs to files.
- `forward`: (available since `v0.1.1`) forward logs to another `prista` instance.
- `fanout`: (available since `v0.1.3`) fan-out logs to other log writers.

(**) It accepts Go-style of datetime format. Therefore, to rotate log file every hour, an example of file name pattern would be `default.log-20060102_15`.

(***) Log file format:
- `tsv`: one line per log entry in the following format `<category-name><tab-character><log-message>`
- `json`: one line per log entry in the following format `{"category":<category-name>, "message": <log-message>}`

(+) Address of the external `prista` instance, either
- `udp://host:port`: forward logs to another `prista` instance via UDP
- or `grpc://host:port`: forward logs to another `prista` instance via gRPC
- or `http://host:port` (or `https://host:port`): forward logs to another `prista` instance via HTTP

(++) Comma-separated list of category names, e.g. `cata,catb,catc`: fan-out logs to 3 other categories named `cata`, `catb` and `catc`.
