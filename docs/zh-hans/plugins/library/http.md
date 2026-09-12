# Http 标准库

`vfox`提供了一个简单的 http 能力，当前支持`Get`、`Head`两种请求类型，以及文件下载。

GET 和 HEAD 请求默认总超时为 30 秒，包含响应体读取。文件下载默认使用独立的 30 分钟总超时。两者的连接超时和等待响应头超时均使用 `plugin.http.timeout`，配置代理时同样生效。超时通过现有的 Lua 错误返回值返回。

可在 `config.yaml` 中覆盖：

```yaml
plugin:
  http:
    timeout: 60s
    downloadTimeout: 45m
```

也可以使用命令行修改：

```shell
vfox config plugin.http.timeout 60s
vfox config plugin.http.downloadTimeout 45m
```

时长支持 `ms`、`s`、`m`、`h` 等单位。未设置的字段或 `0s` 会继承共享配置的对应值，再分别回退到 30 秒和 30 分钟；零值不会禁用超时保护。负数和无效时长会报错。使用 `vfox config --unset plugin.http.timeout` 可恢复该字段的继承行为。

**使用**

```lua
local http = require("http")
--- get 请求, 不要用此请求进行文件下载!!!
local resp, err = http.get({
    url = "https://httpbin.org/json",
    headers = {
      ['Host'] = "localhost"
    }
})
--- 返回参数
assert(err == nil)
assert(resp.status_code == 200)
assert(resp.headers['Content-Type'] == 'application/json')
assert(resp.body == 'xxxxxxxx')

--- head 请求
resp, err = http.head({
    url = "https://httpbin.org/json",
    headers = {
      ['Host'] = "localhost"
    }
})
assert(err == nil)
assert(resp.status_code == 200)
assert(resp.content_length ~= 0)

--- 下载文件, vfox >= 0.4.0
err = http.download_file({
    url = "https://version-fox.github.io/vfox-plugins/index.json",
    headers = {}
}, "/usr/local/file")
assert(err == nil, [[must be nil]] )

```
