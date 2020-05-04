# xlog

Package xlog 是 zap 的简化版本。保留了 zap 大多数功能。
简单粗暴是 golang 的代码哲学，xlog 保持了这一传统。

## 特性

1. 支持日志级别
2. 支持日志名称和预置字段
3. 支持控制台和JSON输出，并可以定制其他格式输出
4. 支持简单的结构化字段
5. 尽量不要使用昂贵的 fmt.Printf 和 reflect

## Examples

### Use Global Logger

``` go
const url = "http://example.com"

// 对于简单的应用可以直接通过使用全局日志函数来记录
// 使用fmt-style
xlog.Infof("Failed to fetch URL: %s", url)

// 使用结构化日志
xlog.Info("Failed to fetch URL.",
	xlog.F("url", url),
	xlog.F("attempt", 3),
	xlog.F("backoff", time.Second),
)
```

### Use xlog.New

``` go
const url = "http://example.com"

l := xlog.New(xlog.NewCore(xlog.NewConsoleEncoder(Llongfile), xlog.Lock(os.Stderr), DebugLevel),
    xlog.AddCaller())
    
// 对于简单的应用可以直接通过使用全局日志函数来记录
// 使用fmt-style
l.Infof("Failed to fetch URL: %s", url)

// 使用结构化日志
l.Info("Failed to fetch URL.",
	xlog.F("url", url),
	xlog.F("attempt", 3),
	xlog.F("backoff", time.Second),
)
```

## 性能

由于简化了 zap 代码，整体性能优于 zap。
