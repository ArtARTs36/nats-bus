package eventhistory

import (
	"fmt"
	natsbus "github.com/artarts36/nats-bus"
	"reflect"
)

type ExtraIDFetcher func(event natsbus.Event) (string, bool)

func emptyExtraIDFetcher(_ natsbus.Event) (string, bool) {
	return "", false
}

func ExtraIDFromPayloadField(field string) ExtraIDFetcher {
	return func(event natsbus.Event) (string, bool) {
		v := reflect.ValueOf(event)
		f := v.FieldByName(field)

		if !f.IsValid() {
			return "", false
		}

		if f.Kind() == reflect.String {
			return v.String(), true
		}

		if f.Kind() == reflect.Int {
			return fmt.Sprintf("%d", v.Int()), true
		}

		return "", false
	}
}
