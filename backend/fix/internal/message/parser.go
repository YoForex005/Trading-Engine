package message

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	SOH = 0x01 // Start of Header delimiter
)

var (
	ErrInvalidMessage    = errors.New("invalid FIX message")
	ErrInvalidChecksum   = errors.New("invalid checksum")
	ErrMissingBeginString = errors.New("missing BeginString (8)")
	ErrMissingBodyLength  = errors.New("missing BodyLength (9)")
	ErrMissingMsgType     = errors.New("missing MsgType (35)")
	ErrMissingChecksum    = errors.New("missing Checksum (10)")
)

// Parser provides fast, zero-allocation FIX message parsing
type Parser struct {
	buffer []byte
	tags   map[int]*TagPosition
}

// TagPosition represents the position of a tag value in the buffer
type TagPosition struct {
	Start int
	End   int
}

// NewParser creates a new FIX message parser
func NewParser() *Parser {
	return &Parser{
		tags: make(map[int]*TagPosition, 50), // Pre-allocate common size
	}
}

// Parse parses a FIX message into tag positions
func (p *Parser) Parse(msg []byte) error {
	p.buffer = msg
	// Clear previous tags
	for k := range p.tags {
		delete(p.tags, k)
	}

	// Validate basic structure
	if len(msg) < 20 {
		return ErrInvalidMessage
	}

	// Parse tags in a single pass
	pos := 0
	for pos < len(msg) {
		// Find tag number
		eqPos := bytes.IndexByte(msg[pos:], '=')
		if eqPos == -1 {
			break
		}
		eqPos += pos

		// Parse tag number
		tagNum, err := parseIntFast(msg[pos:eqPos])
		if err != nil {
			return fmt.Errorf("invalid tag number: %w", err)
		}

		// Find value end (SOH or end of message)
		valueStart := eqPos + 1
		sohPos := bytes.IndexByte(msg[valueStart:], SOH)
		var valueEnd int
		if sohPos == -1 {
			valueEnd = len(msg)
		} else {
			valueEnd = valueStart + sohPos
		}

		// Store tag position
		p.tags[tagNum] = &TagPosition{
			Start: valueStart,
			End:   valueEnd,
		}

		// Move to next tag
		if sohPos == -1 {
			break
		}
		pos = valueEnd + 1
	}

	return p.validateRequiredTags()
}

// validateRequiredTags checks for required FIX header tags
func (p *Parser) validateRequiredTags() error {
	if _, ok := p.tags[8]; !ok {
		return ErrMissingBeginString
	}
	if _, ok := p.tags[9]; !ok {
		return ErrMissingBodyLength
	}
	if _, ok := p.tags[35]; !ok {
		return ErrMissingMsgType
	}
	if _, ok := p.tags[10]; !ok {
		return ErrMissingChecksum
	}
	return nil
}

// GetTag returns the raw bytes for a tag
func (p *Parser) GetTag(tag int) []byte {
	if pos, ok := p.tags[tag]; ok {
		return p.buffer[pos.Start:pos.End]
	}
	return nil
}

// GetTagString returns a tag value as string
func (p *Parser) GetTagString(tag int) string {
	val := p.GetTag(tag)
	if val == nil {
		return ""
	}
	return string(val)
}

// GetTagInt returns a tag value as int
func (p *Parser) GetTagInt(tag int) (int, error) {
	val := p.GetTag(tag)
	if val == nil {
		return 0, fmt.Errorf("tag %d not found", tag)
	}
	return parseIntFast(val)
}

// GetTagFloat returns a tag value as float64
func (p *Parser) GetTagFloat(tag int) (float64, error) {
	val := p.GetTag(tag)
	if val == nil {
		return 0, fmt.Errorf("tag %d not found", tag)
	}
	return strconv.ParseFloat(string(val), 64)
}

// GetTagBool returns a tag value as bool (Y/N)
func (p *Parser) GetTagBool(tag int) (bool, error) {
	val := p.GetTag(tag)
	if val == nil {
		return false, fmt.Errorf("tag %d not found", tag)
	}
	if len(val) == 0 {
		return false, nil
	}
	return val[0] == 'Y' || val[0] == 'y', nil
}

// GetTagTime returns a tag value as time (FIX timestamp format)
func (p *Parser) GetTagTime(tag int) (time.Time, error) {
	val := p.GetTag(tag)
	if val == nil {
		return time.Time{}, fmt.Errorf("tag %d not found", tag)
	}
	return parseFIXTime(string(val))
}

// HasTag checks if a tag exists
func (p *Parser) HasTag(tag int) bool {
	_, ok := p.tags[tag]
	return ok
}

// GetMsgType returns the message type (tag 35)
func (p *Parser) GetMsgType() string {
	return p.GetTagString(35)
}

// GetMsgSeqNum returns the message sequence number (tag 34)
func (p *Parser) GetMsgSeqNum() (int, error) {
	return p.GetTagInt(34)
}

// GetSenderCompID returns the sender comp ID (tag 49)
func (p *Parser) GetSenderCompID() string {
	return p.GetTagString(49)
}

// GetTargetCompID returns the target comp ID (tag 56)
func (p *Parser) GetTargetCompID() string {
	return p.GetTagString(56)
}

// ValidateChecksum validates the message checksum
func (p *Parser) ValidateChecksum() error {
	// Find checksum position in original message
	checksumPos := bytes.LastIndex(p.buffer, []byte("\x0110="))
	if checksumPos == -1 {
		return ErrMissingChecksum
	}

	// Calculate checksum up to but not including "10="
	sum := 0
	for i := 0; i < checksumPos+1; i++ { // Include the SOH before 10=
		sum += int(p.buffer[i])
	}
	calculatedChecksum := sum % 256

	// Get stored checksum
	storedChecksum, err := p.GetTagInt(10)
	if err != nil {
		return fmt.Errorf("invalid checksum tag: %w", err)
	}

	if calculatedChecksum != storedChecksum {
		return fmt.Errorf("%w: calculated=%d, stored=%d", ErrInvalidChecksum, calculatedChecksum, storedChecksum)
	}

	return nil
}

// parseIntFast parses integer without allocations
func parseIntFast(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("empty value")
	}

	negative := false
	start := 0
	if b[0] == '-' {
		negative = true
		start = 1
	}

	result := 0
	for i := start; i < len(b); i++ {
		if b[i] < '0' || b[i] > '9' {
			return 0, fmt.Errorf("invalid digit: %c", b[i])
		}
		result = result*10 + int(b[i]-'0')
	}

	if negative {
		result = -result
	}

	return result, nil
}

// parseFIXTime parses FIX timestamp format (YYYYMMDD-HH:MM:SS or YYYYMMDD-HH:MM:SS.sss)
func parseFIXTime(s string) (time.Time, error) {
	// Format: YYYYMMDD-HH:MM:SS.sss or YYYYMMDD-HH:MM:SS
	if len(s) < 17 {
		return time.Time{}, errors.New("invalid timestamp length")
	}

	layout := "20060102-15:04:05"
	if len(s) > 17 && s[17] == '.' {
		layout = "20060102-15:04:05.000"
	}

	return time.Parse(layout, s)
}

// Builder provides efficient FIX message construction
type Builder struct {
	buf        bytes.Buffer
	bodyStart  int
	seqNum     int
	senderComp string
	targetComp string
}

// NewBuilder creates a new FIX message builder
func NewBuilder(senderCompID, targetCompID string) *Builder {
	return &Builder{
		senderComp: senderCompID,
		targetComp: targetCompID,
	}
}

// BeginMessage starts building a new message
func (b *Builder) BeginMessage(msgType string, seqNum int) {
	b.buf.Reset()
	b.seqNum = seqNum

	// Write BeginString
	b.buf.WriteString("8=FIX.4.4\x01")

	// Reserve space for BodyLength (we'll fill this later)
	b.bodyStart = b.buf.Len()
	b.buf.WriteString("9=00000\x01") // Placeholder

	// Write standard header
	b.AddTag(35, msgType)
	b.AddTag(49, b.senderComp)
	b.AddTag(56, b.targetComp)
	b.AddTag(34, seqNum)
	b.AddTag(52, time.Now().UTC().Format("20060102-15:04:05.000"))
}

// AddTag adds a tag=value pair
func (b *Builder) AddTag(tag int, value interface{}) {
	b.buf.WriteString(strconv.Itoa(tag))
	b.buf.WriteByte('=')

	switch v := value.(type) {
	case string:
		b.buf.WriteString(v)
	case int:
		b.buf.WriteString(strconv.Itoa(v))
	case float64:
		b.buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		if v {
			b.buf.WriteByte('Y')
		} else {
			b.buf.WriteByte('N')
		}
	}

	b.buf.WriteByte(SOH)
}

// Build finalizes and returns the complete message
func (b *Builder) Build() []byte {
	// Get message without header and trailer
	msg := b.buf.Bytes()

	// Calculate body length (from after 9= to before 10=)
	bodyLength := len(msg) - b.bodyStart - 7 // Subtract "9=00000\x01"

	// Update body length
	bodyLengthStr := fmt.Sprintf("9=%d\x01", bodyLength)
	copy(msg[b.bodyStart:], bodyLengthStr)

	// Calculate checksum (entire message up to this point)
	sum := 0
	for _, c := range msg {
		sum += int(c)
	}
	checksum := sum % 256

	// Append checksum
	b.buf.WriteString(fmt.Sprintf("10=%03d\x01", checksum))

	return b.buf.Bytes()
}

// Reset resets the builder for reuse
func (b *Builder) Reset() {
	b.buf.Reset()
	b.bodyStart = 0
}
