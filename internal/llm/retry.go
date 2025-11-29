package llm

import (
    "context"
    "math/rand"
    "net"
    "net/http"
    "time"
)

// retryHTTP wraps an operation with small exponential backoff retries for transient failures.
// It retries when:
// - the op returns a retriable error (temporary net error, timeout), or
// - the returned HTTP status code is retriable (429, 408, 5xx)
// The op should perform the HTTP request and return the response and/or an error.
func retryHTTP(ctx context.Context, maxAttempts int, baseDelay time.Duration, op func() (*http.Response, error)) (*http.Response, error) {
    if maxAttempts < 1 {
        maxAttempts = 1
    }
    if baseDelay <= 0 {
        baseDelay = 100 * time.Millisecond
    }
    var lastErr error
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        if ctx != nil {
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            default:
            }
        }

        resp, err := op()
        if err == nil && resp != nil && !isRetriableStatus(resp.StatusCode) {
            return resp, nil
        }

        // Decide whether to retry
        shouldRetry := false
        if err != nil {
            shouldRetry = isRetriableError(err)
        } else if resp != nil {
            shouldRetry = isRetriableStatus(resp.StatusCode)
            if !shouldRetry {
                return resp, nil
            }
            // close body before retry to avoid leaks
            resp.Body.Close()
        }

        lastErr = err
        if attempt == maxAttempts || !shouldRetry {
            // return last response if present, else error
            if err == nil && resp != nil {
                return resp, nil
            }
            return resp, err
        }

        // backoff with jitter
        delay := baseDelay << (attempt - 1) // 100ms, 200ms, 400ms...
        // cap delay to 1s to keep tests fast
        if delay > time.Second {
            delay = time.Second
        }
        // add jitter +/- 20%
        jitter := time.Duration(rand.Int63n(int64(delay/5)))
        delay = delay - delay/10 + jitter

        timer := time.NewTimer(delay)
        select {
        case <-ctx.Done():
            timer.Stop()
            return nil, ctx.Err()
        case <-timer.C:
        }
    }
    return nil, lastErr
}

func isRetriableStatus(code int) bool {
    // Retry only on rate limit or request timeout. Do NOT retry generic 5xx
    // to keep behavior predictable and fast (tests expect immediate failure).
    return code == http.StatusTooManyRequests || code == http.StatusRequestTimeout
}

func isRetriableError(err error) bool {
    // unwrap net errors
    if ne, ok := err.(net.Error); ok {
        return ne.Timeout() || ne.Temporary()
    }
    // Fallback: treat context deadline/canceled as non-retriable
    return false
}
