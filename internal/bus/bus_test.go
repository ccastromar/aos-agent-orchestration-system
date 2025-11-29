package bus

import (
    "testing"
    "time"
)

func TestBus_SubscribeAndSend_DeliversMessage(t *testing.T) {
    b := New()

    ch := make(chan Message, 1)
    b.Subscribe("worker", ch)

    msg := Message{Type: "hello", Payload: map[string]any{"x": 1}}
    b.Send("worker", msg)

    select {
    case got := <-ch:
        if got.Type != "hello" {
            t.Fatalf("unexpected type: %s", got.Type)
        }
        if got.Payload["x"].(int) != 1 {
            t.Fatalf("unexpected payload: %+v", got.Payload)
        }
    case <-time.After(500 * time.Millisecond):
        t.Fatal("timeout waiting message")
    }
}

func TestBus_SendToUnknown_NoPanicOrBlock(t *testing.T) {
    b := New()
    done := make(chan struct{})
    go func() {
        // No subscriber registered for "nobody"; should safely no-op
        b.Send("nobody", Message{Type: "t"})
        close(done)
    }()

    select {
    case <-done:
        // ok
    case <-time.After(500 * time.Millisecond):
        t.Fatal("send to unknown subscriber blocked")
    }
}
