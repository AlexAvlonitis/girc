package irc

// Numeric replies handled by the client.
const (
	RplWelcome       = "001"
	RplISupport      = "005"
	RplAway          = "301"
	RplUnaway        = "305"
	RplNowAway       = "306"
	RplChannelModeIs = "324"
	RplNoTopic       = "331"
	RplTopic         = "332"
	RplTopicWhoTime  = "333"
	RplNamReply      = "353"
	RplEndOfNames    = "366"
	ErrNicknameInUse = "433"
)

// IsReplyToUser reports whether a numeric is normally a direct answer to a
// command the user typed (WHOIS, WHO, AWAY, INVITE, errors...), as opposed to
// unsolicited server information.
func IsReplyToUser(numeric string) bool {
	if len(numeric) == 3 && (numeric[0] == '4' || numeric[0] == '5') {
		return true // error replies
	}
	switch numeric {
	case RplAway, RplUnaway, RplNowAway,
		"307", "311", "312", "313", "314", "315", "317", "318", "319", "320",
		"330", "338", "341", "352", "369", "378", "379", "671":
		return true
	}
	return false
}
