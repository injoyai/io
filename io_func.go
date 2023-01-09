package io

// multiCloser
// 合并多个Closer , 变成1个Closer
type multiCloser struct {
	closer []Closer
}

func (this *multiCloser) Close() error {
	var err error
	for _, v := range this.closer {
		if er := v.Close(); er != nil {
			err = er
		}
	}
	return err
}

// MultiCloser 多个关闭合并
func MultiCloser(closer ...Closer) Closer {
	return &multiCloser{closer: closer}
}

type publishToWriter struct {
	topic string
	Publisher
}

func (this *publishToWriter) Write(p []byte) (int, error) {
	err := this.Publisher.Publish(this.topic, p)
	return len(p), err
}

func PublisherToWriter(p Publisher, topic string) Writer {
	return &publishToWriter{topic: topic, Publisher: p}
}
