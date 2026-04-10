package service

// fake_chat_scripts.go is intentionally data-heavy so conversation variants are
// easy to review and extend without touching the fake chat flow logic.

type fakeChatTopic string

const (
	fakeTopicName     fakeChatTopic = "name"
	fakeTopicAge      fakeChatTopic = "age"
	fakeTopicLocation fakeChatTopic = "location"
)

var fakeGreetingMessages = []string{
	"Hi nè",
	"hello bạn",
	"hey, có ai ở đây hông",
	"chào cậu nha",
	"helooo",
	"ê chào b",
	"hi bạn ơi",
	"hello hello",
	"yo chào cậu",
	"xin chào, nói chuyện hong",
}

var fakePromptVariants = map[fakeChatTopic][]string{
	fakeTopicName: {
		"b tên gì á",
		"cậu tên gì v",
		"cho mình xin tên để dễ xưng hô nhaa",
		"ê tên gì thế",
		"bạn tên gì vậy",
		"tên em là gì vậy",
		"đánh nhau không",
		"c tên gì á",
		"baby tên gì vậy :>",
		"xưng hô vs bae như nào đây",
	},
	fakeTopicAge: {
		"b nhiêu tuổi r",
		"cậu bao nhiu tuổi á",
		"mấy tuổi vậy b",
		"sn bao nhiêu thế",
		"2k mấy v",
		"tuổi tác sao nè",
		"b thuộc lứa nào á",
		"cho mình hỏi tuổi với",
		"bn tuổi vậy trời",
		"cậu đang tầm bao nhiêu tuổi",
	},
	fakeTopicLocation: {
		"b ở đâu v",
		"cậu ở tỉnh nào á",
		"đang ở khu nào thế",
		"quê ở đâu vậy b",
		"ở thành phố nào nè",
		"quận/huyện nào á",
		"địa chỉ kiểu khu vực thôi, b ở đâu vậy",
		"m đang ở đâu thế",
		"cho t hỏi b ở đâu nha",
		"cậu thuộc team tỉnh nào vậy",
	},
}
