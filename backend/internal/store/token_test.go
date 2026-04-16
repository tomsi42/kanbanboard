package store

import (
	"errors"
	"testing"
)

func TestCreateApiToken_returnsRawToken(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Agent", "agent@test.com")

	token, raw, err := CreateApiToken(db, user.ID, "test-token")
	if err != nil {
		t.Fatalf("CreateApiToken: %v", err)
	}
	if token.ID == "" {
		t.Error("token ID should be set")
	}
	if token.Name != "test-token" {
		t.Errorf("Name = %q, want test-token", token.Name)
	}
	if raw == "" {
		t.Error("raw token should be non-empty")
	}
	if len(raw) != 64 {
		t.Errorf("raw token length = %d, want 64 (hex-encoded 32 bytes)", len(raw))
	}
}

func TestGetUserByToken_found(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Agent", "agent@test.com")

	_, raw, err := CreateApiToken(db, user.ID, "my-token")
	if err != nil {
		t.Fatalf("CreateApiToken: %v", err)
	}

	found, err := GetUserByToken(db, raw)
	if err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("got user %s, want %s", found.ID, user.ID)
	}
}

func TestGetUserByToken_invalidToken(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)

	_, err := GetUserByToken(db, "notavalidtoken")
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("want ErrTokenNotFound, got %v", err)
	}
}

func TestListApiTokens(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Agent", "agent@test.com")

	CreateApiToken(db, user.ID, "token-a")
	CreateApiToken(db, user.ID, "token-b")

	tokens, err := ListApiTokens(db, user.ID)
	if err != nil {
		t.Fatalf("ListApiTokens: %v", err)
	}
	if len(tokens) != 2 {
		t.Errorf("got %d tokens, want 2", len(tokens))
	}
}

func TestDeleteApiToken_removesToken(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Agent", "agent@test.com")

	token, _, err := CreateApiToken(db, user.ID, "to-delete")
	if err != nil {
		t.Fatalf("CreateApiToken: %v", err)
	}

	if err := DeleteApiToken(db, token.ID, user.ID); err != nil {
		t.Fatalf("DeleteApiToken: %v", err)
	}

	tokens, _ := ListApiTokens(db, user.ID)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens after delete, got %d", len(tokens))
	}
}

func TestDeleteApiToken_scopedToUser(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	alice := seedUser(t, db, "Alice", "alice@test.com")
	bob := seedUser(t, db, "Bob", "bob@test.com")

	token, _, err := CreateApiToken(db, alice.ID, "alice-token")
	if err != nil {
		t.Fatalf("CreateApiToken: %v", err)
	}

	// Bob trying to delete Alice's token should get ErrTokenNotFound
	err = DeleteApiToken(db, token.ID, bob.ID)
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("want ErrTokenNotFound when deleting another user's token, got %v", err)
	}

	// Alice's token should still exist
	tokens, _ := ListApiTokens(db, alice.ID)
	if len(tokens) != 1 {
		t.Errorf("expected 1 token, got %d", len(tokens))
	}
}
