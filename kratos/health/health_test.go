package health

import "testing"

func TestToSnake(t *testing.T) {
	got := "GreeterService"
	want := "greeter_service"

	if toSnake(got) != want {
		t.Errorf("toSnake(%q) = %q; want %q", got, toSnake(got), want)
	}
}
