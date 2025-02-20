package form

import (
	"net/url"
	"reflect"

	"github.com/go-playground/form/v4"
	"google.golang.org/protobuf/proto"

	"github.com/go-fox/fox/codec"
)

const (
	// Name is form codec name
	Name = "x-www-form-urlencoded"
	// Null value string
	nullStr = "null"
)

var (
	encoder = form.NewEncoder()
	decoder = form.NewDecoder()
)

// This variable can be replaced with -ldflags like below:
// go build "-ldflags=-X github.com/go-kratos/kratos/v2/encoding/form.tagName=form"
var tagName = "json"

func init() {
	decoder.SetTagName(tagName)
	encoder.SetTagName(tagName)
	codec.RegisterCodec(Codec{encoder: encoder, decoder: decoder})
}

// Codec is form codec
type Codec struct {
	encoder *form.Encoder
	decoder *form.Decoder
}

// Marshal form data
func (c Codec) Marshal(v interface{}) ([]byte, error) {
	var vs url.Values
	var err error
	if m, ok := v.(proto.Message); ok {
		vs, err = EncodeValues(m)
		if err != nil {
			return nil, err
		}
	} else {
		vs, err = c.encoder.Encode(v)
		if err != nil {
			return nil, err
		}
	}
	for k, v := range vs {
		if len(v) == 0 {
			delete(vs, k)
		}
	}
	return []byte(vs.Encode()), nil
}

// Unmarshal form data
func (c Codec) Unmarshal(data []byte, v interface{}) error {
	vs, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if m, ok := v.(proto.Message); ok {
		return DecodeValues(m, vs)
	}
	if m, ok := rv.Interface().(proto.Message); ok {
		return DecodeValues(m, vs)
	}

	return c.decoder.Decode(v, vs)
}

// Name is form codec name
func (Codec) Name() string {
	return Name
}
