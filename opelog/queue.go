package opelog

type QueueV0 interface {
	Enqueue(string) error
	Dequeue() (string, error)
	Close() error
}

type Queue interface {
	Put(string) error
	Get() (string, error)
	Close() error
}
