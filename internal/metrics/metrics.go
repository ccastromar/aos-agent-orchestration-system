package metrics

import (
    "fmt"
    "net/http"
    "sort"
    "strings"
    "sync"
)

// A very small in-process metrics registry that exports Prometheus-like text.
// It supports counters and simple summaries (count/sum), with labeled samples.

type labelsKey string

func makeKey(lbls map[string]string) labelsKey {
    if len(lbls) == 0 {
        return labelsKey("")
    }
    keys := make([]string, 0, len(lbls))
    for k := range lbls { keys = append(keys, k) }
    sort.Strings(keys)
    var b strings.Builder
    for i, k := range keys {
        if i > 0 { b.WriteByte(',') }
        b.WriteString(k)
        b.WriteByte('=')
        // escape quotes
        v := strings.ReplaceAll(lbls[k], "\"", "\\\"")
        b.WriteString("\"")
        b.WriteString(v)
        b.WriteString("\"")
    }
    return labelsKey(b.String())
}

type CounterVec struct {
    Name   string
    Help   string
    mu     sync.RWMutex
    labelNames []string
    values map[labelsKey]float64
}

func NewCounterVec(name, help string, labelNames ...string) *CounterVec {
    return &CounterVec{Name: name, Help: help, labelNames: labelNames, values: make(map[labelsKey]float64)}
}

func (cv *CounterVec) Inc(lbls map[string]string) {
    key := makeKey(lbls)
    cv.mu.Lock()
    cv.values[key] += 1
    cv.mu.Unlock()
}

// SummaryVec stores count and sum; we export metric_count and metric_sum.
type SummaryVec struct {
    Name string
    Help string
    mu   sync.RWMutex
    labelNames []string
    count map[labelsKey]float64
    sum   map[labelsKey]float64
}

func NewSummaryVec(name, help string, labelNames ...string) *SummaryVec {
    return &SummaryVec{Name: name, Help: help, labelNames: labelNames, count: make(map[labelsKey]float64), sum: make(map[labelsKey]float64)}
}

func (sv *SummaryVec) Observe(lbls map[string]string, v float64) {
    key := makeKey(lbls)
    sv.mu.Lock()
    sv.count[key] += 1
    sv.sum[key] += v
    sv.mu.Unlock()
}

// Global metrics we care about
var (
    HTTPRequests = NewCounterVec("aos_http_requests_total", "Total HTTP requests", "method", "path", "status")
    HTTPDuration = NewSummaryVec("aos_http_request_seconds", "HTTP request duration seconds", "method", "path", "status")

    BusMessages  = NewCounterVec("aos_bus_messages_total", "Bus messages by target and result", "target", "result") // result=sent|dropped

    LLMPings     = NewCounterVec("aos_llm_pings_total", "LLM Ping calls", "provider", "outcome") // outcome=ok|error
    LLMChats     = NewCounterVec("aos_llm_chats_total", "LLM Chat calls", "provider", "outcome")
    LLMChatDur   = NewSummaryVec("aos_llm_chat_seconds", "LLM Chat duration seconds", "provider", "outcome")
)

// ServeHTTP exposes all metrics in Prometheus text format.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")

    // helper to dump counter vec
    dumpCounter := func(cv *CounterVec) {
        fmt.Fprintf(w, "# HELP %s %s\n", cv.Name, cv.Help)
        fmt.Fprintf(w, "# TYPE %s counter\n", cv.Name)
        cv.mu.RLock()
        for key, val := range cv.values {
            if key == "" {
                fmt.Fprintf(w, "%s %g\n", cv.Name, val)
            } else {
                fmt.Fprintf(w, "%s{%s} %g\n", cv.Name, key, val)
            }
        }
        cv.mu.RUnlock()
    }

    dumpSummary := func(sv *SummaryVec) {
        // Prometheus summary convention: name_sum and name_count
        fmt.Fprintf(w, "# HELP %s %s\n", sv.Name, sv.Help)
        fmt.Fprintf(w, "# TYPE %s summary\n", sv.Name)
        sv.mu.RLock()
        for key, cnt := range sv.count {
            sum := sv.sum[key]
            if key == "" {
                fmt.Fprintf(w, "%s_sum %g\n", sv.Name, sum)
                fmt.Fprintf(w, "%s_count %g\n", sv.Name, cnt)
            } else {
                fmt.Fprintf(w, "%s_sum{%s} %g\n", sv.Name, key, sum)
                fmt.Fprintf(w, "%s_count{%s} %g\n", sv.Name, key, cnt)
            }
        }
        sv.mu.RUnlock()
    }

    dumpCounter(HTTPRequests)
    dumpSummary(HTTPDuration)
    dumpCounter(BusMessages)
    dumpCounter(LLMPings)
    dumpCounter(LLMChats)
    dumpSummary(LLMChatDur)
}
