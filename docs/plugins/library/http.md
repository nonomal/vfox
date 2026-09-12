# HTTP Library

`vfox` provides a simple HTTP library, currently supporting only GET and HEAD requests and download file. In the Lua script, you can
use `require("http")` to access it. For example:

GET and HEAD requests default to a 30-second total timeout, including reading the response body. File downloads default to a separate 30-minute total timeout. The connection and response-header timeouts use `plugin.http.timeout` for both types of request, including when a proxy is configured. Timeouts are returned through the existing Lua error return value.

Override these settings in `config.yaml`:

```yaml
plugin:
  http:
    timeout: 60s
    downloadTimeout: 45m
```

Or use the CLI:

```shell
vfox config plugin.http.timeout 60s
vfox config plugin.http.downloadTimeout 45m
```

Values use duration units such as `ms`, `s`, `m`, or `h`. Omitted fields and `0s` inherit the corresponding shared configuration value, falling back to 30 seconds and 30 minutes respectively. Zero does not disable timeout protection. Negative and invalid durations are rejected. `vfox config --unset plugin.http.timeout` restores inheritance for that setting.

**Usage**

```lua
local http = require("http")
--- get request, do not use this request to download files!!!
local resp, err = http.get({
    url = "https://httpbin.org/json",
    headers = {
      ['Host'] = "localhost"
    }
})
--- return parameters
assert(err == nil)
assert(resp.status_code == 200)
assert(resp.headers['Content-Type'] == 'application/json')
assert(resp.body == 'xxxxx')

--- head request
resp, err = http.head({
    url = "https://httpbin.org/json",
    headers = {
      ['Host'] = "localhost"
    }
})
assert(err == nil)
assert(resp.status_code == 200)
assert(resp.content_length ~= 0)

--- Download file, vfox >= 0.4.0
err = http.download_file({
    url = "https://version-fox.github.io/vfox-plugins/index.json",
    headers = {}
}, "/usr/local/file")
assert(err == nil, [[must be nil]] )

```
