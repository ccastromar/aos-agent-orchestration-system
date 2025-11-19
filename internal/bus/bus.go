package bus

type Message struct {
    Type    string
    Payload map[string]any
}

type Bus struct {
    subs map[string]chan Message
}

func New() *Bus {
    return &Bus{
        subs: make(map[string]chan Message),
    }
}

func (b *Bus) Subscribe(name string, ch chan Message) {
    b.subs[name] = ch
}

func (b *Bus) Send(target string, msg Message) {
    if ch, ok := b.subs[target]; ok {
        ch <- msg
    }
}
