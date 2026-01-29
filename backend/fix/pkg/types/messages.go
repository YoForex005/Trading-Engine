package types

import (
	"time"
)

// FIX Message Types - Session Level
const (
	MsgTypeLogon          = "A"
	MsgTypeLogout         = "5"
	MsgTypeHeartbeat      = "0"
	MsgTypeTestRequest    = "1"
	MsgTypeResendRequest  = "2"
	MsgTypeReject         = "3"
	MsgTypeSequenceReset  = "4"
	MsgTypeBusinessReject = "j"
)

// FIX Message Types - Trading
const (
	MsgTypeNewOrderSingle           = "D"
	MsgTypeOrderCancelRequest       = "F"
	MsgTypeOrderCancelReplaceRequest = "G"
	MsgTypeOrderCancelReject        = "9"
	MsgTypeOrderStatusRequest       = "H"
	MsgTypeExecutionReport          = "8"
	MsgTypeOrderMassStatusRequest   = "AF"
)

// FIX Message Types - Market Data
const (
	MsgTypeMarketDataRequest         = "V"
	MsgTypeMarketDataSnapshot        = "W"
	MsgTypeMarketDataIncrementalRefresh = "X"
	MsgTypeMarketDataReject          = "Y"
	MsgTypeQuoteRequest              = "R"
	MsgTypeQuote                     = "S"
	MsgTypeMassQuote                 = "i"
)

// FIX Message Types - Position & Trade
const (
	MsgTypeRequestForPositions       = "AN"
	MsgTypeRequestForPositionsAck    = "AO"
	MsgTypePositionReport            = "AP"
	MsgTypeTradeCaptureReportRequest = "AD"
	MsgTypeTradeCaptureReportAck     = "AQ"
	MsgTypeTradeCaptureReport        = "AE"
	MsgTypeTradingSessionStatus      = "h"
)

// Side represents order side
type Side string

const (
	SideBuy  Side = "1"
	SideSell Side = "2"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeMarket          OrderType = "1"
	OrderTypeLimit           OrderType = "2"
	OrderTypeStop            OrderType = "3"
	OrderTypeStopLimit       OrderType = "4"
	OrderTypeMarketOnClose   OrderType = "5"
	OrderTypeWithOrWithout   OrderType = "6"
	OrderTypeLimitOrBetter   OrderType = "7"
	OrderTypeLimitWithOrWithout OrderType = "8"
	OrderTypeOnBasis         OrderType = "9"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusNew              OrderStatus = "0"
	OrderStatusPartiallyFilled  OrderStatus = "1"
	OrderStatusFilled           OrderStatus = "2"
	OrderStatusDoneForDay       OrderStatus = "3"
	OrderStatusCanceled         OrderStatus = "4"
	OrderStatusReplaced         OrderStatus = "5"
	OrderStatusPendingCancel    OrderStatus = "6"
	OrderStatusStopped          OrderStatus = "7"
	OrderStatusRejected         OrderStatus = "8"
	OrderStatusSuspended        OrderStatus = "9"
	OrderStatusPendingNew       OrderStatus = "A"
	OrderStatusCalculated       OrderStatus = "B"
	OrderStatusExpired          OrderStatus = "C"
	OrderStatusAcceptedForBidding OrderStatus = "D"
	OrderStatusPendingReplace   OrderStatus = "E"
)

// ExecType represents execution type
type ExecType string

const (
	ExecTypeNew             ExecType = "0"
	ExecTypePartialFill     ExecType = "1"
	ExecTypeFill            ExecType = "2"
	ExecTypeDoneForDay      ExecType = "3"
	ExecTypeCanceled        ExecType = "4"
	ExecTypeReplaced        ExecType = "5"
	ExecTypePendingCancel   ExecType = "6"
	ExecTypeStopped         ExecType = "7"
	ExecTypeRejected        ExecType = "8"
	ExecTypeSuspended       ExecType = "9"
	ExecTypePendingNew      ExecType = "A"
	ExecTypeCalculated      ExecType = "B"
	ExecTypeExpired         ExecType = "C"
	ExecTypeRestated        ExecType = "D"
	ExecTypePendingReplace  ExecType = "E"
	ExecTypeTrade           ExecType = "F"
)

// TimeInForce represents order time in force
type TimeInForce string

const (
	TimeInForceDay                TimeInForce = "0"
	TimeInForceGoodTillCancel     TimeInForce = "1"
	TimeInForceAtTheOpening       TimeInForce = "2"
	TimeInForceImmediateOrCancel  TimeInForce = "3"
	TimeInForceFillOrKill         TimeInForce = "4"
	TimeInForceGoodTillCrossing   TimeInForce = "5"
	TimeInForceGoodTillDate       TimeInForce = "6"
	TimeInForceAtTheClose         TimeInForce = "7"
)

// MDEntryType represents market data entry type
type MDEntryType string

const (
	MDEntryTypeBid              MDEntryType = "0"
	MDEntryTypeOffer            MDEntryType = "1"
	MDEntryTypeTrade            MDEntryType = "2"
	MDEntryTypeIndexValue       MDEntryType = "3"
	MDEntryTypeOpeningPrice     MDEntryType = "4"
	MDEntryTypeClosingPrice     MDEntryType = "5"
	MDEntryTypeSettlementPrice  MDEntryType = "6"
	MDEntryTypeTradingSessionHighPrice MDEntryType = "7"
	MDEntryTypeTradingSessionLowPrice  MDEntryType = "8"
	MDEntryTypeTradingSessionVWAP      MDEntryType = "9"
)

// MDUpdateAction represents market data update action
type MDUpdateAction string

const (
	MDUpdateActionNew    MDUpdateAction = "0"
	MDUpdateActionChange MDUpdateAction = "1"
	MDUpdateActionDelete MDUpdateAction = "2"
)

// SubscriptionRequestType represents subscription type
type SubscriptionRequestType string

const (
	SubscriptionRequestTypeSnapshot           SubscriptionRequestType = "0"
	SubscriptionRequestTypeSnapshotAndUpdates SubscriptionRequestType = "1"
	SubscriptionRequestTypeUnsubscribe        SubscriptionRequestType = "2"
)

// CancelRejectReason represents order cancel reject reason
type CancelRejectReason string

const (
	CancelRejectReasonTooLateToCancel         CancelRejectReason = "0"
	CancelRejectReasonUnknownOrder            CancelRejectReason = "1"
	CancelRejectReasonBrokerOption            CancelRejectReason = "2"
	CancelRejectReasonOrderAlreadyPendingCancel CancelRejectReason = "3"
)

// SessionRejectReason represents session reject reason
type SessionRejectReason string

const (
	SessionRejectReasonInvalidTagNumber           SessionRejectReason = "0"
	SessionRejectReasonRequiredTagMissing         SessionRejectReason = "1"
	SessionRejectReasonTagNotDefinedForMessageType SessionRejectReason = "2"
	SessionRejectReasonUndefinedTag               SessionRejectReason = "3"
	SessionRejectReasonTagSpecifiedWithoutValue   SessionRejectReason = "4"
	SessionRejectReasonValueIsIncorrect           SessionRejectReason = "5"
	SessionRejectReasonIncorrectDataFormat        SessionRejectReason = "6"
	SessionRejectReasonDecryptionProblem          SessionRejectReason = "7"
	SessionRejectReasonSignatureProblem           SessionRejectReason = "8"
	SessionRejectReasonCompIDProblem              SessionRejectReason = "9"
	SessionRejectReasonSendingTimeAccuracyProblem SessionRejectReason = "10"
	SessionRejectReasonInvalidMsgType             SessionRejectReason = "11"
)

// BusinessRejectReason represents business reject reason
type BusinessRejectReason string

const (
	BusinessRejectReasonOther                          BusinessRejectReason = "0"
	BusinessRejectReasonUnknownID                      BusinessRejectReason = "1"
	BusinessRejectReasonUnknownSecurity                BusinessRejectReason = "2"
	BusinessRejectReasonUnsupportedMessageType         BusinessRejectReason = "3"
	BusinessRejectReasonApplicationNotAvailable        BusinessRejectReason = "4"
	BusinessRejectReasonConditionallyRequiredFieldMissing BusinessRejectReason = "5"
	BusinessRejectReasonNotAuthorized                  BusinessRejectReason = "6"
	BusinessRejectReasonThrottleLimitExceeded          BusinessRejectReason = "7"
)

// FIXMessage represents a base FIX message
type FIXMessage struct {
	MsgType      string
	MsgSeqNum    int
	SenderCompID string
	TargetCompID string
	SendingTime  time.Time
	PossDupFlag  bool
	PossResend   bool
	RawMessage   []byte
	Tags         map[int][]byte
}

// NewOrderSingleRequest represents a new order request
type NewOrderSingleRequest struct {
	ClOrdID     string
	Symbol      string
	Side        Side
	OrderQty    float64
	OrdType     OrderType
	Price       float64
	TimeInForce TimeInForce
	Account     string
	TransactTime time.Time
}

// OrderCancelRequest represents an order cancel request
type OrderCancelRequest struct {
	ClOrdID      string
	OrigClOrdID  string
	Symbol       string
	Side         Side
	TransactTime time.Time
}

// OrderCancelReplaceRequest represents an order modify request
type OrderCancelReplaceRequest struct {
	ClOrdID      string
	OrigClOrdID  string
	Symbol       string
	Side         Side
	OrderQty     float64
	OrdType      OrderType
	Price        float64
	TransactTime time.Time
}

// ExecutionReport represents an execution report
type ExecutionReport struct {
	OrderID      string
	ClOrdID      string
	ExecID       string
	ExecType     ExecType
	OrdStatus    OrderStatus
	Symbol       string
	Side         Side
	OrderQty     float64
	Price        float64
	LastQty      float64
	LastPx       float64
	CumQty       float64
	AvgPx        float64
	LeavesQty    float64
	Text         string
	TransactTime time.Time
}

// MarketDataEntry represents a single market data entry
type MarketDataEntry struct {
	MDEntryType     MDEntryType
	MDEntryPx       float64
	MDEntrySize     float64
	MDEntryPositionNo int
	MDEntryTime     time.Time
	QuoteEntryID    string
}

// MarketDataSnapshot represents a full market data snapshot
type MarketDataSnapshot struct {
	MDReqID    string
	Symbol     string
	NoMDEntries int
	Entries    []MarketDataEntry
	Timestamp  time.Time
}

// MarketDataIncrementalRefresh represents incremental market data update
type MarketDataIncrementalRefresh struct {
	MDReqID     string
	NoMDEntries int
	Updates     []MarketDataUpdate
	Timestamp   time.Time
}

// MarketDataUpdate represents a single market data update
type MarketDataUpdate struct {
	MDUpdateAction  MDUpdateAction
	MDEntryType     MDEntryType
	Symbol          string
	MDEntryPx       float64
	MDEntrySize     float64
	MDEntryPositionNo int
	MDEntryTime     time.Time
}

// MarketDataReject represents a market data subscription reject
type MarketDataReject struct {
	MDReqID   string
	MDReqRejReason string
	Text      string
	Timestamp time.Time
}

// Position represents a position report
type Position struct {
	PosReqID    string
	Symbol      string
	LongQty     float64
	ShortQty    float64
	SettlPrice  float64
	Account     string
	PositionID  string
	Timestamp   time.Time
}

// OrderCancelReject represents an order cancel reject
type OrderCancelReject struct {
	OrderID           string
	ClOrdID           string
	OrigClOrdID       string
	OrdStatus         OrderStatus
	CxlRejResponseTo  string
	CxlRejReason      CancelRejectReason
	Text              string
	TransactTime      time.Time
}

// SessionReject represents a session-level reject
type SessionReject struct {
	RefSeqNum         int
	SessionRejectReason SessionRejectReason
	RefTagID          int
	RefMsgType        string
	Text              string
}

// BusinessMessageReject represents a business-level reject
type BusinessMessageReject struct {
	RefSeqNum            int
	RefMsgType           string
	BusinessRejectRefID  string
	BusinessRejectReason BusinessRejectReason
	Text                 string
}

// ResendRequest represents a resend request
type ResendRequest struct {
	BeginSeqNo int
	EndSeqNo   int
}

// SequenceReset represents a sequence reset message
type SequenceReset struct {
	GapFillFlag bool
	NewSeqNo    int
}

// QuoteRequest represents a quote request
type QuoteRequest struct {
	QuoteReqID string
	Symbol     string
	OrderQty   float64
	Side       Side
	TransactTime time.Time
}

// Quote represents a quote response
type Quote struct {
	QuoteID    string
	QuoteReqID string
	Symbol     string
	BidPx      float64
	OfferPx    float64
	BidSize    float64
	OfferSize  float64
	ValidUntilTime time.Time
}
