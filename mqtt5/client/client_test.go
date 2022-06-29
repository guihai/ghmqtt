package client

import (
	"testing"
)

func TestClient(t *testing.T) {

	cli := NewClient("127.0.0.1", 18883, "tescli", "zf", "zf123")

	cli.Run()

}
