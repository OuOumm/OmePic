package uid

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"
)

const (
	snowflakeWorkerBits         = 5
	snowflakeDatacenterBits     = 5
	snowflakeSequenceBits       = 12
	snowflakeMaxWorkerID        = -1 ^ (-1 << snowflakeWorkerBits)
	snowflakeMaxDatacenterID    = -1 ^ (-1 << snowflakeDatacenterBits)
	snowflakeMaxSequence        = -1 ^ (-1 << snowflakeSequenceBits)
	snowflakeWorkerIDShift      = snowflakeSequenceBits
	snowflakeDatacenterIDShift  = snowflakeSequenceBits + snowflakeWorkerBits
	snowflakeTimestampLeftShift = snowflakeSequenceBits + snowflakeWorkerBits + snowflakeDatacenterBits
	snowflakeEpochMillis        = int64(1704067200000) // 2024-01-01T00:00:00Z

	base62Alphabet      = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base62Base          = 62
	shortIDLength       = 8
	sidByteLength       = 8
	maxPublicUIDLength  = 30
	publicUIDHeadLength = 1
)

var (
	ErrInvalidToken  = errors.New("invalid uid token")
	ErrInvalidPrefix = errors.New("invalid uid prefix")
)

type SIDGenerator interface {
	Next() (int64, error)
}

type Codec struct {
	prefix       string
	secret       []byte
	sidGenerator SIDGenerator
}

type Decoded struct {
	Prefix  string
	ShortID string
	RawUID  string
}

type SnowflakeGenerator struct {
	mu            sync.Mutex
	workerID      int64
	datacenterID  int64
	lastTimestamp int64
	sequence      int64
	now           func() time.Time
}

func NewCodec(prefix string, secret string) (*Codec, error) {
	return NewCodecWithGenerator(prefix, secret, NewSnowflakeGenerator(0, 0), nil)
}

func NewCodecWithGenerator(prefix string, secret string, sidGenerator SIDGenerator, _ io.Reader) (*Codec, error) {
	normalizedPrefix, err := normalizePrefix(prefix)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("uid encryption key must not be empty")
	}
	if sidGenerator == nil {
		return nil, errors.New("sid generator must not be nil")
	}
	maxBodyLength := maxBase62LengthForByteCount(base64.StdEncoding.EncodedLen(len(normalizedPrefix) + sidByteLength))
	if publicUIDHeadLength+maxBodyLength > maxPublicUIDLength {
		return nil, fmt.Errorf("uid prefix is too long to fit within %d characters", maxPublicUIDLength)
	}

	return &Codec{
		prefix:       normalizedPrefix,
		secret:       []byte(secret),
		sidGenerator: sidGenerator,
	}, nil
}

func (c *Codec) Generate() (string, error) {
	sid, err := c.sidGenerator.Next()
	if err != nil {
		return "", err
	}
	if sid <= 0 {
		return "", ErrInvalidToken
	}

	payload := append(sidToBytes(sid), []byte(c.prefix)...)
	xorOffset := xorOffsetForSID(sid)
	xorBytes := applyXORWithOffset(payload, c.secret, xorOffset)
	base64Token := base64.StdEncoding.EncodeToString(xorBytes)
	publicUID := string(base62Alphabet[xorOffset]) + base62EncodeBytes([]byte(base64Token))

	if len(publicUID) > maxPublicUIDLength {
		return "", fmt.Errorf("generated uid exceeds %d characters", maxPublicUIDLength)
	}

	return publicUID, nil
}

func (c *Codec) Validate(token string) error {
	_, err := c.Decode(token)
	return err
}

func (c *Codec) Decode(token string) (Decoded, error) {
	token = strings.TrimSpace(token)
	if len(token) <= publicUIDHeadLength || len(token) > maxPublicUIDLength {
		return Decoded{}, ErrInvalidToken
	}

	xorOffset := strings.IndexByte(base62Alphabet, token[0])
	if xorOffset < 0 {
		return Decoded{}, ErrInvalidToken
	}

	base64Bytes, err := base62DecodeToBytes(token[publicUIDHeadLength:])
	if err != nil || len(base64Bytes) == 0 {
		return Decoded{}, ErrInvalidToken
	}

	xorBytes, err := base64.StdEncoding.DecodeString(string(base64Bytes))
	if err != nil || len(xorBytes) == 0 {
		return Decoded{}, ErrInvalidToken
	}

	payload := string(applyXORWithOffset(xorBytes, c.secret, xorOffset))
	if len(payload) <= sidByteLength {
		return Decoded{}, ErrInvalidToken
	}

	sid, ok := bytesToSID([]byte(payload[:sidByteLength]))
	if !ok {
		return Decoded{}, ErrInvalidToken
	}
	shortID := normalizeShortID(base62EncodeInt64(sid))
	prefix := payload[sidByteLength:]
	if prefix != c.prefix {
		return Decoded{}, ErrInvalidPrefix
	}

	rawUID := prefix + shortID

	return Decoded{
		Prefix:  c.prefix,
		ShortID: shortID,
		RawUID:  rawUID,
	}, nil
}

func normalizePrefix(prefix string) (string, error) {
	value := strings.TrimSpace(prefix)
	value = strings.TrimRight(value, "_")
	if value == "" {
		return "", errors.New("uid prefix must not be empty")
	}
	return value + "_", nil
}

func normalizeShortID(encoded string) string {
	switch {
	case len(encoded) < shortIDLength:
		return strings.Repeat("0", shortIDLength-len(encoded)) + encoded
	case len(encoded) > shortIDLength:
		return encoded[len(encoded)-shortIDLength:]
	default:
		return encoded
	}
}

func isValidShortID(value string) bool {
	if len(value) != shortIDLength {
		return false
	}
	for i := 0; i < len(value); i++ {
		if !strings.ContainsRune(base62Alphabet, rune(value[i])) {
			return false
		}
	}
	return true
}

func applyXORWithOffset(payload []byte, secret []byte, offset int) []byte {
	result := make([]byte, len(payload))
	for i := range payload {
		result[i] = payload[i] ^ secret[(i+offset)%len(secret)]
	}
	return result
}

func xorOffsetForSID(sid int64) int {
	offset := sid % base62Base
	if offset < 0 {
		offset += base62Base
	}
	return int(offset)
}

func sidToBytes(value int64) []byte {
	output := make([]byte, sidByteLength)
	binary.LittleEndian.PutUint64(output, uint64(value))
	return output
}

func bytesToSID(value []byte) (int64, bool) {
	if len(value) != sidByteLength {
		return 0, false
	}

	raw := binary.LittleEndian.Uint64(value)
	if raw == 0 || raw > uint64(math.MaxInt64) {
		return 0, false
	}

	return int64(raw), true
}

func base62EncodeInt64(value int64) string {
	return base62EncodeBigInt(big.NewInt(value))
}

func base62EncodeBytes(value []byte) string {
	number := new(big.Int).SetBytes(value)
	return base62EncodeBigInt(number)
}

func base62EncodeBigInt(value *big.Int) string {
	if value == nil || value.Sign() == 0 {
		return "0"
	}

	number := new(big.Int).Set(value)
	base := big.NewInt(base62Base)
	remainder := new(big.Int)
	var output []byte

	for number.Sign() > 0 {
		number.DivMod(number, base, remainder)
		output = append(output, base62Alphabet[remainder.Int64()])
	}

	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}

	return string(output)
}

func base62DecodeToBytes(value string) ([]byte, error) {
	number := big.NewInt(0)
	base := big.NewInt(base62Base)

	for i := 0; i < len(value); i++ {
		index := strings.IndexByte(base62Alphabet, value[i])
		if index < 0 {
			return nil, ErrInvalidToken
		}

		number.Mul(number, base)
		number.Add(number, big.NewInt(int64(index)))
	}

	return number.Bytes(), nil
}

func maxBase62LengthForByteCount(byteCount int) int {
	if byteCount <= 0 {
		return 1
	}

	limit := new(big.Int).Lsh(big.NewInt(1), uint(byteCount*8))
	value := big.NewInt(base62Base)
	length := 1

	for value.Cmp(limit) < 0 {
		value.Mul(value, big.NewInt(base62Base))
		length++
	}

	return length
}

func NewSnowflakeGenerator(workerID int64, datacenterID int64) *SnowflakeGenerator {
	if workerID < 0 || workerID > snowflakeMaxWorkerID {
		workerID = 0
	}
	if datacenterID < 0 || datacenterID > snowflakeMaxDatacenterID {
		datacenterID = 0
	}

	return &SnowflakeGenerator{
		workerID:     workerID,
		datacenterID: datacenterID,
		now:          time.Now,
	}
}

func (g *SnowflakeGenerator) Next() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	nowMillis := g.now().UTC().UnixMilli()
	if nowMillis < snowflakeEpochMillis {
		return 0, fmt.Errorf("snowflake clock is before epoch")
	}
	if nowMillis < g.lastTimestamp {
		return 0, fmt.Errorf("clock moved backwards")
	}

	if nowMillis == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & snowflakeMaxSequence
		if g.sequence == 0 {
			nowMillis = g.waitNextMillis(g.lastTimestamp)
		}
	} else {
		g.sequence = 0
	}

	g.lastTimestamp = nowMillis

	sid := ((nowMillis - snowflakeEpochMillis) << snowflakeTimestampLeftShift) |
		(g.datacenterID << snowflakeDatacenterIDShift) |
		(g.workerID << snowflakeWorkerIDShift) |
		g.sequence

	return sid, nil
}

func (g *SnowflakeGenerator) waitNextMillis(last int64) int64 {
	nowMillis := g.now().UTC().UnixMilli()
	for nowMillis <= last {
		time.Sleep(time.Millisecond)
		nowMillis = g.now().UTC().UnixMilli()
	}
	return nowMillis
}
