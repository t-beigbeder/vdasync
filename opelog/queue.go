package opelog

type QueueV0 interface {
	Enqueue(string) error
	Dequeue() (string, error)
	Close() error
}

type Queue interface {
	Put([]byte) error
	Get() ([]byte, error)
	Close() error
}
