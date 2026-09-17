package matching

const (
	ModeMixed          = "mixed"
	ModeOppositeSex    = "opposite_sex"
	QueueDisplayHidden  = "hidden"
	QueueDisplayMessage = "message"
	QueueDisplayCount   = "count"
)

const DefaultQueueMessageVI = "Anochat còn mới nên đôi khi việc tìm người trò chuyện sẽ mất một chút thời gian. Cảm ơn bạn đã chờ nhé!"
const DefaultQueueMessageEN = "Anochat is still new, so finding someone may take a little time. Thanks for waiting!"

func ValidMode(mode string) bool {
	return mode == ModeMixed || mode == ModeOppositeSex
}

func ValidQueueDisplayMode(mode string) bool {
	return mode == QueueDisplayHidden || mode == QueueDisplayMessage || mode == QueueDisplayCount
}

func ValidRematchCooldownSeconds(seconds int) bool {
	switch seconds {
	case 0, 6*60*60, 12*60*60, 24*60*60, 7*24*60*60:
		return true
	default:
		return false
	}
}

type Settings struct {
	DefaultMode            string
	AllowUserChoice        bool
	RematchCooldownSeconds int
	QueueDisplayMode       string
	QueueCountMinimum      int
	QueueMessageVI         string
	QueueMessageEN         string
}

func (s Settings) EffectiveMode(preference *string) string {
	if s.AllowUserChoice && preference != nil && ValidMode(*preference) {
		return *preference
	}
	return s.DefaultMode
}
