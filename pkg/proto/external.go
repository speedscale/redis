package proto

import (
	"io"

	redis "github.com/go-redis/redis/v8/internal/proto"
)

// The creators of this redis client never intended the protocol reader to be made public.
// Unfortunately, we need this specific code to be externalized for use in our proxy.
// To minimize rebasing headaches we'll just put a wrapper around the bits we need.
// Yes, your Computer Science professor would not approve of circumventing the go 'internal'
// package protections but she isn't here to reject this MR. Just be glad I didn't copy/paste
// it instead.
type Reader struct {
	r *redis.Reader
}

// NewReader is a simple wrapper around the internal go-redis Reader.
func NewReader(rd io.Reader) *Reader {
	return &Reader{
		r: redis.NewReader(rd),
	}
}

func (r *Reader) ReadReply() (interface{}, error) {
	return r.r.ReadReply(sliceParser)
}

// sliceParser implements proto.MultiBulkParse. This is copy/pasted from command.go so that we can make
// some slight modifications if needed.
func sliceParser(rd *redis.Reader, n int64) (interface{}, error) {
	vals := make([]interface{}, n)
	for i := 0; i < len(vals); i++ {
		v, err := rd.ReadReply(sliceParser)
		if err != nil {
			if err == redis.Nil {
				vals[i] = nil
				continue
			}
			if err, ok := err.(redis.RedisError); ok {
				vals[i] = err
				continue
			}
			return nil, err
		}
		vals[i] = v
	}
	return vals, nil
}
