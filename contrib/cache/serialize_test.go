package cache

import (
	"reflect"
	"testing"
)

func TestSerializer(t *testing.T) {
	type args struct {
		v any
		w any
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "int",
			args: args{
				v: int(1),
				w: int(1),
			},
		},
		{
			name: "string",
			args: args{
				v: "hello",
				w: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Serializer{}
			got, err := s.Marshal(tt.args.v)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}
			of := reflect.TypeOf(tt.args.v)
			pof := reflect.PointerTo(reflect.TypeOf(tt.args.v))
			value := reflect.New(pof)
			value.Set(reflect.New(of))
			v := value.Elem().Interface()
			if err := s.Unmarshal(got, v); err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}
			if !reflect.DeepEqual(tt.args.v, tt.args.w) {
				t.Errorf("Marshal() got = %v, want %v", tt.args.v, tt.args.w)
			}
		})
	}
}

func TestSerualizer2(t *testing.T) {
	s := &Serializer{}
	var b string = "hello"
	var bres string
	marshal, err := s.Marshal(b)
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
		return
	}
	if err := s.Unmarshal(marshal, &bres); err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}
}
