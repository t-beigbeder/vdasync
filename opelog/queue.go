package opelog

type Queue interface {
	Enqueue(string) error
	Dequeue() (string, error)
	Close() error
}