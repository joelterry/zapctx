# zapctx [![Go Reference](https://pkg.go.dev/badge/github.com/joelterry/zapctx.svg)](https://pkg.go.dev/github.com/joelterry/zapctx)

Have you ever used zap.Logger.With() and thought to yourself, "all these nonessential scoped values sure make me think of the context package"? Me too. The behavior of zap.Logger.With() is so similar to that of context.WithValue() that I imagine some people opt to forgo their loggers' contextual functionality and redundantly log context values instead. Or, if one chooses to use their logger as a contextual object, they have to either store it as a context value or pass it everywhere as a second argument after the standardized context.Context.

The online wisdom regarding loggers and contexts (meaning, a quick Google search yielding this [blog post](https://gogoapps.io/blog/passing-loggers-in-go-golang-logging-best-practices/) that references this other [blog post](https://dave.cheney.net/2017/01/26/context-is-for-cancelation)) suggests that loggers should be explicit dependencies (struct fields) rather than implicit dependencies (context values).

What if there was another way? What if the *Logger* was the **Context**... and the **Context** was the *Logger*?

![meme](meme.jpg)

The [context](https://pkg.go.dev/context) package consists of three separate Context implementations (*cancelCtx, *timerCtx, *valueCtx) performing distinct scoped jobs. What's one more: *loggerCtx? Though I do believe there's a strong enough correlation between logging and contexts to warrant this additional functionality, it isn't a good idea due to the size and unwieldiness of the hypothetical Logger interface: logging hasn't been upgraded to a standard library interface the way the file system has.

Having personally settled on Uber's [zap](https://github.com/uber-go/zap) package for logging, I wrote this package to try this idea out myself. So far I've been enjoying it! It behaves like a superset of the context package, meaning one can replace

```go
import "context"
```

with 

```go
import context "github.com/joelterry/zapctx"
```

and start logging from their ctx objects immediately.

## Usage

This package is probably best used as a blanket replacement for context in an isolated codebase. However, here's a more complicated example to demonstrate interoperability with standard context. Imagine three packages with respective functions A, B, C. A uses zapctx, B uses standard context, and C uses both. There is a chain of function calls A->B->C:

```go
package A

import "https://github.com/uber-go/zap"
import context "github.com/joelterry/zapctx"
import "B"
   
func A() {
    logger, _ := zap.NewProduction()
    ctx := context.WithLogger(logger).With(zap.String("key", "value"))
    ctx.Info("logging from a context!")
    B.B(ctx)
}
```

```go 
package B

import "context"
import "C"

func B(ctx context.Context) {
    ctx2, cancel := context.WithCancel(ctx) 
    C.C(ctx2)
    cancel()
}
```

```go
package C

import "context"
import "github.com/joelterry/zapctx"

func C(ctx context.Context) {
    zctx := zapctx.Logger(ctx)
    zctx.Info("a zapctx.Context can be recovered from a context.Context chain")
    <-zctx.Done()
}
```

## Prior Art

Before writing this package, I came across two others also named "zapctx", both addressing the same issue. [github.com/juju/zaputil/zapctx](https://pkg.go.dev/github.com/juju/zaputil/zapctx) in particular was close to what I wanted: it stores loggers in the context value chain, but doesn't allow you to log from context directly. Instead, you would have to unwrap a logger within every function that logs, or use logging functions that take (context.Context, string, ...zap.Field). Though this package might seem intrusive, I like that for the most part it merges APIs in lieue of making additions or changes (less boilerplate).