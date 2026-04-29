package uid

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

type staticSIDGenerator struct {
	values []int64
}

func (g *staticSIDGenerator) Next() (int64, error) {
	if len(g.values) == 0 {
		return 0, errors.New("no sid values queued")
	}
	value := g.values[0]
	g.values = g.values[1:]
	return value, nil
}

func TestCodecGenerateAndDecodeRoundTrip(t *testing.T) {
	codec, err := NewCodecWithGenerator(
		"omeo_",
		"test-secret",
		&staticSIDGenerator{values: []int64{123456789}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator returned error: %v", err)
	}

	token, err := codec.Generate()
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if got := len(token); got > maxPublicUIDLength {
		t.Fatalf("expected token length <= %d, got %d", maxPublicUIDLength, got)
	}

	decoded, err := codec.Decode(token)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if decoded.Prefix != "omeo_" {
		t.Fatalf("expected prefix omeo_, got %q", decoded.Prefix)
	}
	if decoded.ShortID != normalizeShortID(base62EncodeInt64(123456789)) {
		t.Fatalf("expected short id %q, got %q", normalizeShortID(base62EncodeInt64(123456789)), decoded.ShortID)
	}
	if decoded.RawUID != "omeo_"+decoded.ShortID {
		t.Fatalf("expected raw uid %q, got %q", "omeo_"+decoded.ShortID, decoded.RawUID)
	}
}

func TestCodecNormalizesPrefixWithSingleSeparator(t *testing.T) {
	withoutUnderscore, err := NewCodecWithGenerator("omeo", "test-secret", &staticSIDGenerator{values: []int64{1}}, nil)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator without underscore returned error: %v", err)
	}
	withUnderscore, err := NewCodecWithGenerator("omeo_", "test-secret", &staticSIDGenerator{values: []int64{1}}, nil)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator with underscore returned error: %v", err)
	}

	first, err := withoutUnderscore.Generate()
	if err != nil {
		t.Fatalf("Generate without underscore returned error: %v", err)
	}
	second, err := withUnderscore.Generate()
	if err != nil {
		t.Fatalf("Generate with underscore returned error: %v", err)
	}
	if first != second {
		t.Fatalf("expected normalized prefixes to generate identical tokens, got %q and %q", first, second)
	}

	decoded, err := withUnderscore.Decode(first)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if decoded.RawUID != "omeo_00000001" {
		t.Fatalf("expected normalized raw uid omeo_00000001, got %q", decoded.RawUID)
	}
}

func TestCodecUsesEightCharacterShortIDPaddingAndTruncation(t *testing.T) {
	codec, err := NewCodecWithGenerator(
		"omeo_",
		"test-secret",
		&staticSIDGenerator{values: []int64{1, 218340105584896}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator returned error: %v", err)
	}

	first, err := codec.Generate()
	if err != nil {
		t.Fatalf("first Generate returned error: %v", err)
	}
	firstDecoded, err := codec.Decode(first)
	if err != nil {
		t.Fatalf("first Decode returned error: %v", err)
	}
	if firstDecoded.ShortID != "00000001" {
		t.Fatalf("expected padded short id 00000001, got %q", firstDecoded.ShortID)
	}

	second, err := codec.Generate()
	if err != nil {
		t.Fatalf("second Generate returned error: %v", err)
	}
	secondDecoded, err := codec.Decode(second)
	if err != nil {
		t.Fatalf("second Decode returned error: %v", err)
	}
	if secondDecoded.ShortID != "00000000" {
		t.Fatalf("expected truncated short id 00000000, got %q", secondDecoded.ShortID)
	}
}

func TestCodecVariesPublicUIDHeadAcrossGeneratedValues(t *testing.T) {
	codec, err := NewCodecWithGenerator(
		"omeo_",
		"test-secret",
		&staticSIDGenerator{values: []int64{1, 2}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator returned error: %v", err)
	}

	first, err := codec.Generate()
	if err != nil {
		t.Fatalf("first Generate returned error: %v", err)
	}
	second, err := codec.Generate()
	if err != nil {
		t.Fatalf("second Generate returned error: %v", err)
	}

	if first == second {
		t.Fatalf("expected generated public UIDs to differ across calls, got %q", first)
	}
	if first[0] == second[0] {
		t.Fatalf("expected varying public UID head for different generated values, got %q and %q", first, second)
	}

	firstDecoded, err := codec.Decode(first)
	if err != nil {
		t.Fatalf("first Decode returned error: %v", err)
	}
	secondDecoded, err := codec.Decode(second)
	if err != nil {
		t.Fatalf("second Decode returned error: %v", err)
	}

	if firstDecoded.RawUID != "omeo_00000001" {
		t.Fatalf("expected first raw uid %q, got %q", "omeo_00000001", firstDecoded.RawUID)
	}
	if secondDecoded.RawUID != "omeo_00000002" {
		t.Fatalf("expected second raw uid %q, got %q", "omeo_00000002", secondDecoded.RawUID)
	}
}

func TestCodecVariesCiphertextWhenXOROffsetRepeats(t *testing.T) {
	codec, err := NewCodecWithGenerator(
		"omeo_",
		"test-secret",
		&staticSIDGenerator{values: []int64{2, 64}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator returned error: %v", err)
	}

	first, err := codec.Generate()
	if err != nil {
		t.Fatalf("first Generate returned error: %v", err)
	}
	second, err := codec.Generate()
	if err != nil {
		t.Fatalf("second Generate returned error: %v", err)
	}

	if first[0] != second[0] {
		t.Fatalf("expected repeated XOR offset to keep the same public UID head, got %q and %q", first, second)
	}
	if first[1:5] == second[1:5] {
		t.Fatalf("expected leading public UID body segment to vary even when XOR offset repeats, got %q and %q", first[1:5], second[1:5])
	}

	firstCiphertext := mustDecodeCiphertext(t, first)
	secondCiphertext := mustDecodeCiphertext(t, second)
	if len(firstCiphertext) < 2 || len(secondCiphertext) < 2 {
		t.Fatal("expected ciphertext with at least two bytes")
	}
	if string(firstCiphertext[:2]) == string(secondCiphertext[:2]) {
		t.Fatalf("expected leading ciphertext segment to vary even when XOR offset repeats, got %q and %q", string(firstCiphertext[:2]), string(secondCiphertext[:2]))
	}
}

func TestCodecPreservesShortIDAcrossSameOffsetValues(t *testing.T) {
	codec, err := NewCodecWithGenerator(
		"omeo_",
		"test-secret",
		&staticSIDGenerator{values: []int64{2, 64, 126}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator returned error: %v", err)
	}

	expected := []string{
		"omeo_00000002",
		"omeo_00000012",
		"omeo_00000022",
	}
	for _, want := range expected {
		token, err := codec.Generate()
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}

		decoded, err := codec.Decode(token)
		if err != nil {
			t.Fatalf("Decode returned error: %v", err)
		}
		if decoded.RawUID != want {
			t.Fatalf("expected raw uid %q, got %q", want, decoded.RawUID)
		}
	}
}

func TestCodecRejectsMalformedCiphertext(t *testing.T) {
	codec, err := NewCodec("omeo_", "test-secret")
	if err != nil {
		t.Fatalf("NewCodec returned error: %v", err)
	}

	if err := codec.Validate("not-a-token!"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestCodecRejectsPrefixMismatchAfterDecrypt(t *testing.T) {
	sourceCodec, err := NewCodecWithGenerator(
		"omeo_",
		"shared-secret",
		&staticSIDGenerator{values: []int64{999}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator source returned error: %v", err)
	}
	targetCodec, err := NewCodec("other_", "shared-secret")
	if err != nil {
		t.Fatalf("NewCodec target returned error: %v", err)
	}

	token, err := sourceCodec.Generate()
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if err := targetCodec.Validate(token); !errors.Is(err, ErrInvalidPrefix) {
		t.Fatalf("expected ErrInvalidPrefix, got %v", err)
	}
}

func TestCodecRejectsTamperedSuffix(t *testing.T) {
	codec, err := NewCodecWithGenerator(
		"omeo_",
		"test-secret",
		&staticSIDGenerator{values: []int64{321}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodecWithGenerator returned error: %v", err)
	}

	token, err := codec.Generate()
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	tampered := token[:len(token)-1] + "A"
	if strings.HasSuffix(token, "A") {
		tampered = token[:len(token)-1] + "B"
	}

	if err := codec.Validate(tampered); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestNewCodecRejectsPrefixesThatExceedLengthBudget(t *testing.T) {
	_, err := NewCodec("prefix-too-long", "test-secret")
	if err == nil {
		t.Fatal("expected constructor to reject overlong prefix")
	}
}

func mustDecodeCiphertext(t *testing.T, token string) []byte {
	t.Helper()

	base64Bytes, err := base62DecodeToBytes(token[publicUIDHeadLength:])
	if err != nil {
		t.Fatalf("base62DecodeToBytes returned error: %v", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(string(base64Bytes))
	if err != nil {
		t.Fatalf("base64 decode returned error: %v", err)
	}
	return ciphertext
}
