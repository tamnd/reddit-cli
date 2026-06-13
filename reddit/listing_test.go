package reddit

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestDecodeListing(t *testing.T) {
	const body = `{"kind":"Listing","data":{"after":"t3_z","dist":2,"children":[
		{"kind":"t3","data":{"id":"a"}},
		{"kind":"t3","data":{"id":"b"}}
	]}}`
	ld, err := decodeListing([]byte(body))
	if err != nil {
		t.Fatalf("decodeListing: %v", err)
	}
	if ld.After != "t3_z" || len(ld.Children) != 2 {
		t.Errorf("after/children = %q/%d", ld.After, len(ld.Children))
	}
}

func TestDecodeListingWrongKind(t *testing.T) {
	if _, err := decodeListing([]byte(`{"kind":"t3","data":{}}`)); err == nil {
		t.Error("expected an error for a non-Listing kind")
	}
}

func TestClassifyErrorBody(t *testing.T) {
	cases := []struct {
		name string
		code int
		body string
		want error
	}{
		{"private", 403, `{"reason":"private"}`, ErrPrivate},
		{"banned", 404, `{"reason":"banned"}`, ErrBanned},
		{"quarantined", 403, `{"reason":"quarantined"}`, ErrBlocked},
		{"plain 404", 404, `{}`, ErrNotFound},
		{"plain 403", 403, `<html>blocked</html>`, ErrBlocked},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := classifyErrorBody(c.code, []byte(c.body))
			if !errors.Is(err, c.want) {
				t.Errorf("got %v, want %v", err, c.want)
			}
		})
	}
}

func TestFlattenTree(t *testing.T) {
	// A top comment with one reply and a "more" stub at the top level.
	const replies = `{"kind":"Listing","data":{"children":[
		{"kind":"t1","data":{"id":"c2","body":"reply"}}
	]}}`
	top := map[string]any{"id": "c1", "body": "top", "replies": json.RawMessage(replies)}
	topData, _ := json.Marshal(top)
	children := []thingEnvelope{
		{Kind: "t1", Data: topData},
		{Kind: "more", Data: json.RawMessage(`{"count":5,"children":["x1","x2"]}`)},
	}
	var out []Comment
	var pending []moreData
	flattenTree(children, &out, &pending)
	if len(out) != 2 {
		t.Fatalf("expected 2 flattened comments (top + reply), got %d", len(out))
	}
	if out[0].CommentID != "c1" || out[1].CommentID != "c2" {
		t.Errorf("order = %q,%q", out[0].CommentID, out[1].CommentID)
	}
	if out[0].ReplyCount != 1 {
		t.Errorf("top reply count = %d, want 1", out[0].ReplyCount)
	}
	if len(pending) != 1 || pending[0].Count != 5 {
		t.Errorf("pending more = %+v", pending)
	}
}

func TestDecodeMoreChildren(t *testing.T) {
	const body = `{"json":{"data":{"things":[
		{"kind":"t1","data":{"id":"m1","body":"x"}},
		{"kind":"more","data":{"count":2,"children":["y1"]}}
	]}}}`
	things, err := decodeMoreChildren([]byte(body))
	if err != nil {
		t.Fatalf("decodeMoreChildren: %v", err)
	}
	if len(things) != 2 {
		t.Fatalf("expected 2 things, got %d", len(things))
	}
}
