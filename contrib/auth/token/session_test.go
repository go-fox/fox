package token

import (
	"sync"
	"testing"
)

func Test_session_AddTokenSign(t *testing.T) {
	type fields struct {
		serializeData *SerializeData
		repository    Repository
		lock          sync.RWMutex
	}
	type args struct {
		sign *Sign
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test add token sign",
			fields: fields{
				serializeData: &SerializeData{
					SignList: SignList{
						{
							Value:  "tokenValue",
							Device: "test",
						},
					},
				},
				repository: nil,
				lock:       sync.RWMutex{},
			},
			args: args{
				sign: &Sign{
					Value:  "tokenValue",
					Device: "deviceValue",
					Extra: map[string]interface{}{
						"test": "测试",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &session{
				serializeData: tt.fields.serializeData,
				repository:    tt.fields.repository,
				lock:          tt.fields.lock,
			}
			if err := s.AddTokenSign(tt.args.sign); (err != nil) != tt.wantErr {
				t.Errorf("addTokenSign() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
