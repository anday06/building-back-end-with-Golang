package service

import (
	"task-management-api/pkg/password"
	"testing"
)

func TestPasswordHashAndCompare(t *testing.T) {
	hash, err := password.Hash("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	if !password.Compare(hash, "correct-password") {
		t.Fatal("expected password to match")
	}
	if password.Compare(hash, "wrong-password") {
		t.Fatal("expected password mismatch")
	}
}
