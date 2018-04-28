# Nursery

[![Build Status](https://travis-ci.org/JamesOwenHall/nursery.svg?branch=master)](https://travis-ci.org/JamesOwenHall/nursery)

Nursery provides a control flow mechanism for guaranteeing that multiple goroutines will exit before continuing execution. This library was inspired by the concept of nurseries discussed here:

https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/

### Examples

```go
// Call Supervise() to create a scope for concurrency to exist. Supervise() will block
// while goroutines run.
err := nursery.Supervise(context.Background(), func(n *nursery.N) {
    // Spawn many goroutines.
    for _, url := range urls {
        url := url
        n.Go(func() error {
            return crawl(url)
        })
    }
})

// At this point, all goroutines have exited. The value of err is the first non-nil
// error returned by a goroutine.
if err != nil {
    return err
}
```
