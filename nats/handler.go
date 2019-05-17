package nats

type Handler struct {
	Name string
	Func func(data []byte)
}
